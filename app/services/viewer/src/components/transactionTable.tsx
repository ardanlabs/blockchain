import React, { FC } from 'react'
import { transaction } from '../../types/index.d'

interface BlockTableProps {
  transaction: transaction
}

const TransactionTable: FC<BlockTableProps> = ({ transaction }) => {
  return (
    <div>
      <table>
        <tbody>
          <tr>
            <td className="key">Chain ID:</td>
            <td className="value">{transaction.chain_id}</td>
            <td className="key">Nonce:</td>
            <td className="value">{transaction.nonce}</td>
            <td className="key">Value:</td>
            <td className="value">{transaction.value}</td>
          </tr>
          <tr>
            <td className="key">To:</td>
            <td colSpan={5} className="value">
              {transaction.to}
            </td>
          </tr>
          <tr>
            <td className="key">Tip:</td>
            <td className="value">{transaction.tip}</td>
            <td className="key">V:</td>
            <td className="value">{transaction.v}</td>
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
            <td className="key">Data:</td>
            <td colSpan={5} className="value">
            </td>
          </tr>
          <tr>
            <td className="key">R:</td>
            <td colSpan={5} className="value">
              {transaction.r}
            </td>
          </tr>
          <tr>
            <td className="key">S:</td>
            <td colSpan={5} className="value">
              {transaction.s}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  )
}

export default TransactionTable
