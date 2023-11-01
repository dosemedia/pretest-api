import { Express, Request, Response } from 'express'
import Controller from './Controller'
import _ from 'lodash'
import sendVerificationEmail, { SendVerificationEmailJobData } from '../queues/sendVerificationEmail'
import sendPasswordResetEmail, { SendPasswordResetEmailJobData } from '../queues/sendPasswordResetEmail'
import sendPasswordChangedEmail, { SendPasswordChangedEmailJobData } from '../queues/sendPasswordChangedEmail'
import sendUserDestroyedEmail, { SendUserDestroyedEmailJobData } from '../queues/sendUserDestroyedEmail'
import cleanupDestroyedUserFiles, { CleanupDestroyedUserFilesJobData } from '../queues/cleanupDestroyedUserFiles'
import bcrypt from 'bcryptjs'
import prisma from '../database'
import { generateTokenForUser } from '../auth'
import cache from '../cache'
import { v4 as uuidv4 } from 'uuid';

class ActionsController implements Controller {

  async getUserForRequest(req: Request) {
    const userId = req.body.session_variables['x-hasura-user-id']

    const user = await prisma.users.findUnique({
      where: {
        id: userId,
      },
    })
    return user
  }

  async wrapErrorHandler(code: () => Promise<any>, res: Response) {
    try {
      await code()
    } catch (error) {
      console.error(error)
      if (error instanceof Error) {
        return res.status(400).send({ message: error.message })
      } else {
        return res.status(400).send({ message: error })
      }
    }
  }

  startup (app: Express) {
    app.post('/hasura/actions/register', async (req: Request, res: Response) => {
      await this.wrapErrorHandler(async () => {
        let email = req.body.input.email
        if (!email) {
          return res.status(400).send({ message: 'Email is required.' })
        }
        // Always handle emails in lowercase on the backend
        email = email.toLowerCase()
    
        let password = req.body.input.password
        if (password.length < 5) {
          return res.status(400).send({ message: 'Password must be at least 5 characters long.' })
        }
        const hashedPassword = await bcrypt.hash(password, 10)

        // Check for existing user so we can send a nicer error message then unique key constraint
        const existing = await prisma.users.findUnique({
          where: {
            email,
          },
        })
        if (existing) {
          throw new Error(`User with email ${email} already exists.`)
        }
    
        // Create the user in the database (unique key constraint will cause error if user already exists)
        const user = await prisma.users.create({
          data: {
            email,
            hashed_password: hashedPassword
          },
        })

        // Create auth token for user
        const token = generateTokenForUser(user)
    
        // Send verification email
        await sendVerificationEmail.add('send verification email for user id ' + user.id, new SendVerificationEmailJobData(user.id), {
          attempts: 3,
          backoff: {
            type: 'exponential',
            delay: 10000
          }
        })
    
        return res.send({ token, id: user.id })
      }, res)
    })

    app.post('/hasura/actions/login', async (req: Request, res: Response) => {
      await this.wrapErrorHandler(async () => {
        let email = req.body.input.email
        if (!email) {
          return res.status(400).send({ message: 'Email is required.' })
        }
        // Always handle emails in lowercase on the backend
        email = email.toLowerCase()
    
        let password = req.body.input.password

        const user = await prisma.users.findUnique({
          where: {
            email
          },
        })
        if (!user) {
          throw new Error('User not found!')
        }

        const passwordMatches = await bcrypt.compare(password, user.hashed_password)
        if (passwordMatches) {
          const token = generateTokenForUser(user)
          return res.send({ token, id: user.id })
        } else {
          return res.status(400).send({ message: 'Email or password did not match.' })
        }
      }, res)
    })

    app.post('/hasura/actions/resendVerificationEmail', async (req: Request, res: Response) => {
      await this.wrapErrorHandler(async () => {
        const user = await this.getUserForRequest(req)
        if (!user) {
          throw new Error('User not found!')
        }

        if (user.email_verified) {
          throw new Error('Email already verified!')
        }
        
        await sendVerificationEmail.add('send verification email for user id ' + user.id, new SendVerificationEmailJobData(user.id), {
          attempts: 3,
          backoff: {
            type: 'exponential',
            delay: 10000
          }
        })

        res.json(true)
      }, res)
    })

    app.post('/hasura/actions/verifyEmail', async (req: Request, res: Response) => {
      await this.wrapErrorHandler(async () => {
        const code = req.body.input.code
        if (!code) {
          return res.status(400).send({ message: 'code is required.' })
        }

        const user = await prisma.users.findFirstOrThrow({
          where: {
            email_verification_code: code,
          },
        })
        
        await prisma.users.update({
          where: {
            id: user.id,
          },
          data: {
            email_verification_code: null,
            email_verified: true
          },
        })
        
        res.json(true)
      }, res)
    })

    app.post('/hasura/actions/sendPasswordResetEmail', async (req: Request, res: Response) => {
      await this.wrapErrorHandler(async () => {
        let email = req.body.input.email
        if (!email) {
          return res.status(400).send({ message: 'Email is required.' })
        }
        // Always handle emails in lowercase on the backend
        email = email.toLowerCase()

        const user = await prisma.users.findFirst({
          where: {
            email,
          },
        })
        if (!user) {
          throw new Error('Email not found')
        }
        
        await sendPasswordResetEmail.add('send password reset email for user id ' + user.id, new SendPasswordResetEmailJobData(user.id), {
          attempts: 3,
          backoff: {
            type: 'exponential',
            delay: 10000
          }
        })
        res.json(true)
      }, res)
    })

    app.post('/hasura/actions/resetPassword', async (req: Request, res: Response) => {
      await this.wrapErrorHandler(async () => {
        let email = req.body.input.email
        if (!email) {
          return res.status(400).send({ message: 'Email is required.' })
        }
        // Always handle emails in lowercase on the backend
        email = email.toLowerCase()

        const password = req.body.input.newPassword
        if (password.length < 5) {
          return res.status(400).send({ message: 'Password must be at least 5 characters long.' })
        }
        const hashedPassword = await bcrypt.hash(password, 10)

        let code = req.body.input.code
        // Don't let anybody reset passwords without a full code!
        if (code.length < 6) {
          return res.status(400).send({ message: 'Code must be at least 6 characters long.' })
        }

        const user = await prisma.users.findFirst({
          where: {
            password_reset_code: code,
            email: email
          },
        })
        if (!user) {
          throw new Error('User not found')
        }

        await prisma.users.update({
          where: {
            id: user.id,
          },
          data: {
            password_reset_code: null,
            hashed_password: hashedPassword,
            password_at: new Date()
          },
        })

        // Clear cached auth tokens for this user
        await cache.flushPrefix(user.id + ':')
        
        await sendPasswordChangedEmail.add('send password changed email for user id ' + user.id, new SendPasswordChangedEmailJobData(user.id), {
          attempts: 3,
          backoff: {
            type: 'exponential',
            delay: 10000
          }
        })
        res.json(true)
      }, res)
    })

    app.post('/hasura/actions/changePassword', async (req: Request, res: Response) => {
      await this.wrapErrorHandler(async () => {
        const user = await this.getUserForRequest(req)
        if (!user) {
          throw new Error('User not found!')
        }

        const oldPassword = req.body.input.oldPassword
        const passwordMatches = await bcrypt.compare(oldPassword, user.hashed_password)
        if (!passwordMatches) {
          throw new Error('Old password did not match')
        }

        const newPassword = req.body.input.newPassword
        if (newPassword.length < 5) {
          return res.status(400).send({ message: 'Password must be at least 5 characters long.' })
        }
        const hashedPassword = await bcrypt.hash(newPassword, 10)

        await prisma.users.update({
          where: {
            id: user.id,
          },
          data: {
            password_reset_code: null,
            hashed_password: hashedPassword,
            password_at: new Date()
          },
        })

        // Clear cached auth tokens for this user
        await cache.flushPrefix(user.id + ':')
        
        await sendPasswordChangedEmail.add('send password changed email for user id ' + user.id, new SendPasswordChangedEmailJobData(user.id), {
          attempts: 3,
          backoff: {
            type: 'exponential',
            delay: 10000
          }
        })
        res.json(true)
      }, res)
    })

    app.post('/hasura/actions/changeEmail', async (req: Request, res: Response) => {
      await this.wrapErrorHandler(async () => {
        const user = await this.getUserForRequest(req)
        if (!user) {
          throw new Error('User not found!')
        }

        const password = req.body.input.password
        const passwordMatches = await bcrypt.compare(password, user.hashed_password)
        if (!passwordMatches) {
          throw new Error('Password did not match')
        }

        const newEmail = req.body.input.newEmail
        if (!newEmail) {
          return res.status(400).send({ message: 'Email is required.' })
        }
        if (newEmail === user.email) {
          return res.status(400).send({ message: 'Cannot use same email.' })
        }

        await prisma.users.update({
          where: {
            id: user.id,
          },
          data: {
            email: newEmail,
            email_verified: false
          },
        })

        await sendVerificationEmail.add('send verification email for user id ' + user.id, new SendVerificationEmailJobData(user.id), {
          attempts: 3,
          backoff: {
            type: 'exponential',
            delay: 10000
          }
        })
        res.json(true)
      }, res)
    })

    app.post('/hasura/actions/destroyUser', async (req: Request, res: Response) => {
      await this.wrapErrorHandler(async () => {
        const user = await this.getUserForRequest(req)
        if (!user) {
          throw new Error('User not found!')
        }

        const password = req.body.input.password
        const passwordMatches = await bcrypt.compare(password, user.hashed_password)
        if (!passwordMatches) {
          throw new Error('Password did not match')
        }

        // TODO - in a transaction - cleanup user data and files (maybe send to a bg job)

        const email = user.email
        const userId = user.id

        await prisma.users.delete({
          where: {
            id: user.id,
          },
        })

        // Clear cached auth tokens for this user
        await cache.flushPrefix(user.id + ':')

        await sendUserDestroyedEmail.add('send destruction email for user id ' + user.id, new SendUserDestroyedEmailJobData(email), {
          attempts: 3,
          backoff: {
            type: 'exponential',
            delay: 10000
          }
        })
        await cleanupDestroyedUserFiles.add('cleanup destroyed user files for user id ' + user.id, new CleanupDestroyedUserFilesJobData(userId), {
          attempts: 3,
          backoff: {
            type: 'exponential',
            delay: 10000
          }
        })
        res.json(true)
      }, res)
    })

    app.post('/hasura/actions/createProject', async (req: Request, res: Response) => {
      await this.wrapErrorHandler(async () => {
        const user = await this.getUserForRequest(req)
        if (!user) {
          throw new Error('User not found!')
        }
        const team_id = req.body.input.team_id
        const name = req.body.input.name
        const team = await prisma.teams.findUnique({ where: {
          id: team_id
        }})
        if (!team) {
          throw new Error('Team not found!')
        }
        
        const uuid = uuidv4()
        const [project, team_project] = await prisma.$transaction([
          prisma.projects.create({ data: { name: name, id: uuid }}),
          prisma.teams_projects.create({ data: { project_id: uuid, team_id: team.id }})
        ])

        res.json({
          id: project.id,
          name: project.name,
          team_id: team_project.team_id
        })
      }, res)
    })

    app.post('/hasura/actions/createTeam', async (req: Request, res: Response) => {
      await this.wrapErrorHandler(async () => {
        const user = await this.getUserForRequest(req)
        if (!user) {
          throw new Error('User not found!')
        }
        
        const name = req.body.input.name
        if (!name) {
          throw new Error('Name is required!')
        }

        const team_id = uuidv4()

        await prisma.$transaction([
          prisma.teams.create({ data: { name, id: team_id }}),
          prisma.teams_users.create({ data: { user_id: user.id, team_id, role: 'admin' } })
        ])
        res.json(true)
      }, res)
    })

    app.post('/hasura/actions/joinTeam', async (req: Request, res: Response) => {
      await this.wrapErrorHandler(async () => {
        const user = await this.getUserForRequest(req)
        if (!user) {
          throw new Error('User not found!')
        }

        const teamId = req.body.input.teamId

        // Find invitation
        const invitation = await prisma.invitations.findFirst({
          where: {
            team_id: teamId,
            email: user.email,
          },
        })
        if (!invitation) {
          return res.status(400).send({ message: 'Invitation not found.' })
        }

        // Double check that the team still exists
        const team = await prisma.teams.findUnique({
          where: {
            id: teamId,
          },
        })
        if (!team) {
          return res.status(400).send({ message: 'Team not found.' })
        }

        // User exists, team exists, and invitation exsists => create membership (and delete invitation)
        await prisma.$transaction([
          prisma.teams_users.create({
            data: {
              user_id: user.id,
              team_id: teamId
            },
          }),
          prisma.invitations.deleteMany({
            where: {
              team_id: teamId,
              email: user.email,
            },
          })
        ])

        res.json(true)
      }, res)
    })

    app.post('/hasura/actions/leaveTeam', async (req: Request, res: Response) => {
      await this.wrapErrorHandler(async () => {
        const user = await this.getUserForRequest(req)
        if (!user) {
          throw new Error('User not found!')
        }

        const teamId = req.body.input.teamId

        // TODO - (in a transaction) peform any cleanup needed when user leaves a team

        console.log('~~ delete teams_users', user.id, teamId)

        await prisma.teams_users.deleteMany({
          where: {
            user_id: user.id,
            team_id: teamId
          },
        })

        res.json(true)
      }, res)
    })
  }

  async shutdown () {
    // No shutdown actions for this controller
  }
}

const controller = new ActionsController()

export default controller
