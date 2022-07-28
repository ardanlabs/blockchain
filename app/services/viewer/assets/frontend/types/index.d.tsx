export interface transaction {
  chain_id: number,
  data: null,
  gas_price: number,
  gas_units: number,
  nonce: number,
  r: number,
  s: number,
  timestamp: number,
  tip: number,
  to: string,
  v: number,
  value: number,
}
export interface block {
  block: {
    number: number,
    prev_block_hash: string,
    timestamp: number,
    beneficiary: string,
    difficulty: number,
    mining_reward: number
    state_root: string,
    trans_root: string,
    nonce: number,
  },
  hash: string,
  trans: transaction[]
}

export interface node {
  active: boolean,
  wsUrl: string,
  httpUrl: string,
  port: number,
  nodeID: number,
  accountID: string, // soon to be account type
  state: 'Connecting...' | 'Mining...' | 'Connected' | 'Connection open',
  blocks: block[],
  successfull: boolean,
}