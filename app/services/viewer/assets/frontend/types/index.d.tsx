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
export type block = {
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