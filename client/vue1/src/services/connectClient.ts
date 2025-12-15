import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { BounceBot } from '../gen/bouncebot_pb'

// Use current hostname so it works from other devices on the network
const serverHost = window.location.hostname || 'localhost'

const transport = createConnectTransport({
  baseUrl: `http://${serverHost}:8080`,
})

export const bounceBotClient = createClient(BounceBot, transport)
