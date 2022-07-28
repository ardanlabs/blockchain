import React, { FC } from 'react'
import { block } from '../../types/index.d'
import BlockTable from './blockTable'

interface BlockProps {
  nodeID: number,
  blockNumber: number,
  block: block,
  successfullNode: boolean,
}

const Block: FC<BlockProps> = ({
  nodeID,
  blockNumber,
  successfullNode,
  block,
}) => {
  const id = `Block-${nodeID}-${blockNumber}`
  let extraClass: string = ''
  if (successfullNode) {
    extraClass = ' mine'
  }

  const classes = `block${extraClass}`
  return (
    <div id={id} className={classes}>
      <BlockTable {...{ block }} />
    </div>
  )
}

export default Block
