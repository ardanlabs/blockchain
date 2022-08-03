import React, { FC } from 'react'
import { block } from '../../types/index.d'

interface BlockTableProps {
  block: block
}

const BlockTable: FC<BlockTableProps> = ({ block }) => {
  return (
    <table>
      <tbody>
        <tr>
          <td className="key">Own Hash:</td>
          <td colSpan={5} className="value">
            {block.hash}
          </td>
        </tr>
        <tr>
          <td className="key">Previous Hash:</td>
          <td colSpan={5} className="value">
            {block.block.prev_block_hash}
          </td>
        </tr>
        <tr>
          <td className="key">Block Number:</td>
          <td className="value">{block.block.number}</td>
          <td className="key">Mining Difficulty:</td>
          <td className="value">{block.block.difficulty}</td>
          <td className="key">Mining Reward:</td>
          <td className="value">{block.block.mining_reward}</td>
        </tr>
        <tr>
          <td className="key">Timestamp:</td>
          <td className="value">{block.block.timestamp}</td>
          <td className="key">No. of Transactions:</td>
          <td className="value">{block.trans.length}</td>
          <td className="key">Nonce:</td>
          <td className="value">{block.block.nonce}</td>
        </tr>
        <tr>
          <td className="key">Beneficiary:</td>
          <td colSpan={5} className="value">
            {block.block.beneficiary}
          </td>
        </tr>
        <tr>
          <td className="key">Transaction Root:</td>
          <td colSpan={5} className="value">
            {block.block.trans_root}
          </td>
        </tr>
        <tr>
          <td className="key">State Root:</td>
          <td colSpan={5} className="value">
            {block.block.state_root}
          </td>
        </tr>
      </tbody>
    </table>
  )
}

export default BlockTable
