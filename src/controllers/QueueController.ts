import { ExpressAdapter, createBullBoard, BullMQAdapter } from '@bull-board/express'
import { Express } from 'express'
import Controller from './Controller'
import basicAuth from 'express-basic-auth'
import { queues, workers } from '../queues'

const adapters: Array<BullMQAdapter> = []
for (const queue of queues) {
  adapters.push(new BullMQAdapter(queue))
}

const serverAdapter = new ExpressAdapter();
serverAdapter.setBasePath('/admin/queues');
createBullBoard({
  queues: adapters,
  serverAdapter,
})

class QueueController implements Controller {
  startup (app: Express) {
    app.use('/admin/queues',
      basicAuth({
        users: { [process.env.QUEUE_UI_USER || 'bullboard'] : process.env.QUEUE_UI_PASS || 'supersecret' },
        challenge: true
      }),
      serverAdapter.getRouter())
  }

  async shutdown () {
    for (const worker of workers) {
      await worker.close()
    }
  }
}

const controller = new QueueController()

export default controller
