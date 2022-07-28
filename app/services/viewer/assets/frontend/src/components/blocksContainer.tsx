import React, { FC } from 'react'
import Block from './block'
import { block } from '../../types/index.d'

interface BlocksContainerProps {
  blocksProp: block[]
  nodeID: number
  successfullNode: boolean
}

const BlocksContainer: FC<BlocksContainerProps> = (props: {
  blocksProp: block[],
  nodeID: number,
  successfullNode: boolean,
}) => {
  const { blocksProp, nodeID, successfullNode } = props
  let blocks: JSX.Element[] = []
  blocksProp.forEach((block) =>{
    blocks.push(<Block key={block.block.number} {...{nodeID, blockNumber: block.block.number, successfullNode, block }} />)
  })
  return <div id={`blocks-${nodeID}`}>{blocks}</div>
}

export default BlocksContainer
