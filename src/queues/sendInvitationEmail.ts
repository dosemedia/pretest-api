// Adding a new queue/worker?  Make sure to add them to QueueController.ts!

import { Queue, Worker, Job } from 'bullmq'
import queueRedisConnection from './queueRedisConnection'
import prisma from '../database'
import fs from 'fs-extra'
import Handlebars from 'handlebars'
import { sendEmail } from '../email'

const queueName = 'sendInvitationEmail'

const queue = new Queue(queueName, { connection: queueRedisConnection })

export class SendInvitationEmailJobData {
  teamId = ''
  email = ''
  constructor(teamId: string, email: string) {
    this.teamId = teamId
    this.email = email
  }
}

export const worker = new Worker(queueName, async job => {
  const teamId = (job.data as SendInvitationEmailJobData).teamId
  const email = (job.data as SendInvitationEmailJobData).email

  // Find the invitation
  const invitation = await prisma.invitations.findFirst({
    where: {
      team_id: teamId,
      email: email,
    },
  })
  if (!invitation) {
    throw new Error('Invitation not found!')
  }

  const team = await prisma.teams.findUnique({
    where: {
      id: teamId,
    },
  })
  if (!team) {
    throw new Error('Team not found!')
  }


  // Send an email with a link containing the code

  const appName = process.env.APP_NAME || 'App Name'
  const acceptInvitationUrl = process.env.WEB_BASE_URL + '/team/' + teamId + '/join'

  const emailHtml = await fs.readFile('./src/emails/build_production/team-invitation.html', 'utf8')
  const template = Handlebars.compile(emailHtml)
  const templateData = { appName, acceptInvitationUrl, teamName: team.name }
  const templateFilled = template(templateData)

  const subject = appName + ' invitation to join : ' + team.name
  const textMessage = `Please visit ${acceptInvitationUrl} to accept your invitation to ${team.name}.`
  
  await sendEmail(email, subject, textMessage, templateFilled)
}, { connection: queueRedisConnection })

worker.on('completed', (job: Job, returnvalue: any) => {
  console.log(`~~ ${queueName} completed`, job.id, returnvalue)
})

worker.on('failed', (job: any, error: Error) => {
  console.error(`~~ ${queueName} failed`, job.id, error)
})

export default queue