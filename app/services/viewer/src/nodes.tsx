import { node } from '../types/index.d'
const nodes: node[] = [
  // The active property sets if the fetch for this node is executed.
  {
    active: true,
    wsUrl: 'ws://localhost:8080/v1/events',
    httpUrl: 'http://localhost:9080/v1/node/block/list/1/latest',
    port: 8080,
    nodeID: 1,
    accountID: '0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8',
    state: 'Connecting...',
    blocks: [],
    successfull: false,
  },
{
    active: true,
    wsUrl: 'ws://localhost:8280/v1/events',
    httpUrl: 'http://localhost:9280/v1/node/block/list/1/latest',
    port: 8280,
    nodeID: 2,
    accountID: '0xb8Ee4c7ac4ca3269fEc242780D7D960bd6272a61',
    state: 'Connecting...',
    blocks: [],
    successfull: false,
  },
{
    active: true,
    wsUrl: 'ws://localhost:8380/v1/events',
    httpUrl: 'http://localhost:9380/v1/node/block/list/1/latest',
    port: 8380,
    nodeID: 3,
    accountID: '0x616c90073c78ac073D89E750836401a92B16dE7e',
    state: 'Connecting...',
    blocks: [],
    successfull: false,
  },
]

export default nodes