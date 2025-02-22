import { Queue, Worker } from 'bullmq'
import sendVerificationEmail, { worker as sendVerificationEmailWorker } from './sendVerificationEmail'
import sendPasswordResetEmail, { worker as sendPasswordResetEmailWorker } from './sendPasswordResetEmail'
import sendPasswordChangedEmail, { worker as sendPasswordChangedEmailWorker } from './sendPasswordChangedEmail'
import sendUserDestroyedEmail, { worker as sendUserDestroyedEmailWorker } from './sendUserDestroyedEmail'
import cleanupDestroyedUserFiles, { worker as cleanupDestroyedUserFilesWorker } from './cleanupDestroyedUserFiles'
import notifyContactFormSubmission, { worker as notifyContactFormSubmissionWorker } from './notifyContactFormSubmission'
import sendInvitationEmail, { worker as sendInvitationEmailWorker } from './sendInvitationEmail'

export const queues: Array<Queue> = [
  sendVerificationEmail,
  sendPasswordResetEmail,
  sendPasswordChangedEmail,
  sendUserDestroyedEmail,
  cleanupDestroyedUserFiles,
  notifyContactFormSubmission,
  sendInvitationEmail
]
export const workers: Array<Worker> = [
  sendVerificationEmailWorker,
  sendPasswordResetEmailWorker,
  sendPasswordChangedEmailWorker,
  sendUserDestroyedEmailWorker,
  cleanupDestroyedUserFilesWorker,
  notifyContactFormSubmissionWorker,
  sendInvitationEmailWorker
]