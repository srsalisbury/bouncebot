import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { BounceBot } from '../gen/bouncebot_pb'
import { config } from '../config'

const transport = createConnectTransport({
  baseUrl: config.httpBaseUrl,
})

export const bounceBotClient = createClient(BounceBot, transport)
