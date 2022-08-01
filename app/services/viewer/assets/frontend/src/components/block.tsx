import React, { FC, useState } from 'react'
import { block } from '../../types/index.d'
import BlockTable from './blockTable'

interface BlockProps {
  nodeID: number,
  blockNumber: number,
  block: block,
  successfullNode: boolean,
  clickHandler: Function,
}

const Block: FC<BlockProps> = ({
  nodeID,
  blockNumber,
  successfullNode,
  block,
  clickHandler,
}) => {
  const [showTransactions, setShowTransactions] = useState(false);
  const id = `Block-${nodeID}-${blockNumber}`
  let extraClass: string = ''

  if (successfullNode) {
    extraClass = ' mine'
  }

  const classes = `block${extraClass}`

  return (
    <div id={id} className={classes} onClick={() => clickHandler(block)}>
      <BlockTable {...{ block }} />
    </div>
  )
}

export default Block
