import React, { FC } from 'react'
import { mempoolTransaction } from '../../types/index.d'

interface mempoolTableProps {
  transaction: mempoolTransaction
}

const MempoolTable: FC<mempoolTableProps> = (mempoolTableProps) => {
  const { transaction } = mempoolTableProps
  return (
    <div>
      <table>
        <tr>
          <td className="key">From:</td>
          <td colSpan={5} className="value">
            {transaction.from}
          </td>
        </tr>
        <tr>
          <td className="key">From Name:</td>
          <td colSpan={5} className="value">
            {transaction.from_name}
          </td>
        </tr>
        <tr>
          <td className="key">To:</td>
          <td colSpan={5} className="value">
            {transaction.to}
          </td>
        </tr>
        <tr>
          <td className="key">To Name:</td>
          <td colSpan={5} className="value">
            {transaction.to_name}
          </td>
        </tr>
        <tr>
          <td className="key">Chain ID:</td>
          <td className="value">{transaction.chain_id}</td>
          <td className="key">Nonce:</td>
          <td className="value">{transaction.nonce}</td>
          <td className="key">Value:</td>
          <td className="value">{transaction.value}</td>
        </tr>
        <tr>
          <td className="key">Tip:</td>
          <td className="value">{transaction.tip}</td>
          <td className="key">Data:</td>
          {/* <td className="value">{transaction.data}</td> */}
          <td className="key">Timestamp:</td>
          <td className="value">{transaction.timestamp}</td>
        </tr>
        <tr>
          <td className="key">Gas Price:</td>
          <td colSpan={2} className="value">
            {transaction.gas_price}
          </td>
          <td className="key">Gas Units:</td>
          <td colSpan={2} className="value">
            {transaction.gas_units}
          </td>
        </tr>
        <tr>
          <td className="key">Sig:</td>
          <td colSpan={5} className="value">
            {transaction.sig}
          </td>
        </tr>
        <tr>
          <td className="key">Proof:</td>
          <td colSpan={5} className="value">
            {transaction.proof}
          </td>
        </tr>
        <tr>
          <td className="key">Proof Order:</td>
          <td colSpan={5} className="value">
            {transaction.proof_order}
          </td>
        </tr>
      </table>
    </div>
  )
}

export default MempoolTable
