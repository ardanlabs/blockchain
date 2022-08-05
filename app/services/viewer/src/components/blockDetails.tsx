import React, { FC } from 'react'
import { block } from '../../types/index.d'
import BlockTable from './blockTable'

interface BlockDetailsProps {
  nodeID: number,
  blockNumber: number,
  block: block,
  successfullNode: boolean,
}

const BlockDetails: FC<BlockDetailsProps> = (BlockDetailsProps) => {
  const { nodeID, blockNumber, block, successfullNode } = BlockDetailsProps
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

export default BlockDetails
