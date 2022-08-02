import React, { FC } from 'react'
import Block from './block'
import { block } from '../../types/index.d'

interface BlocksContainerProps {
  blocksProp: block[]
  nodeID: number
  successfullNode: boolean
  clickHandler: Function,
}

const BlocksContainer: FC<BlocksContainerProps> = (props: {
  blocksProp: block[],
  nodeID: number,
  successfullNode: boolean,
  clickHandler: Function,
}) => {
  const { blocksProp, nodeID, successfullNode, clickHandler } = props
  let blocks: JSX.Element[] = []
  const addedBlocks: Set<string> = new Set()
  blocksProp.forEach((block) =>{
    if (!addedBlocks.has(`${nodeID}-${block.block.number}`))
      blocks.push(<Block key={block.block.number} {...{nodeID, blockNumber: block.block.number, successfullNode, block, clickHandler }} />)
    addedBlocks.add(`${nodeID}-${block.block.number}`)
  })
  return <div id={`blocks-${nodeID}`}>{blocks}</div>
}

export default BlocksContainer
