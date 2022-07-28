import React, { FC } from 'react'
import { block } from '../../types/index.d'
import Block from './block'

interface BlockContainerProps {
  nodeID: string,
  blockNumber: number,
  block: block,
  successfullNode: boolean,
}

const BlockContainer: FC<BlockContainerProps> = ({ nodeID, blockNumber, successfullNode, block }) => {
  const id = `Block-${nodeID}-${blockNumber}`
  let extraClass: string = ''
  if (successfullNode) {
    extraClass = ' mine'
  }

  const classes = `block${extraClass}`
  return (
    <div id={id} className={classes}>
      <Block {...{nodeID, blockNumber, successfullNode, block }} />
    </div>
  )
}

export default BlockContainer
