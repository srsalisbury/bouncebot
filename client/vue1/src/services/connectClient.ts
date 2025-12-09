import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { BounceBot } from '../gen/bouncebot_pb'

const transport = createConnectTransport({
  baseUrl: 'http://localhost:8080',
})

export const bounceBotClient = createClient(BounceBot, transport)
