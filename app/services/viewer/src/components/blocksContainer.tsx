import React, { FC } from 'react'
import Block from './block'
import { block } from '../../types/index.d'
import LinkIcon from './icons/linkIcon'

interface BlocksContainerProps {
  blocksProp: block[]
  nodeID: number
  successfullNode: boolean
}

const BlocksContainer: FC<BlocksContainerProps> = (BlocksContainerProps) => {
  const { blocksProp, nodeID, successfullNode } = BlocksContainerProps
  let blocks: JSX.Element[] = []
  const addedBlocks: Set<string> = new Set()
  for (let i = 0; i < blocksProp.length; i++) {
    const block = blocksProp[i];
    if (!addedBlocks.has(`${nodeID}-${block.block.number}`)) {
      blocks.push(
        <div key={block.block.number} className="d-flex align-items-center justify-content-center">
          <Block
            {...{
              nodeID,
              blockNumber: block.block.number,
              successfullNode,
              block,
            }}
          />
          <LinkIcon key={`${block.block.number}-Link`} />
        </div>
      )
    }
    addedBlocks.add(`${nodeID}-${block.block.number}`)
  }
  return <div id={`blocks-${nodeID}`} className="my-3">{blocks}</div>
}

export default BlocksContainer
