import React, { FC } from 'react'
import Block from './block'
import { block } from '../../types/index.d'
import LinkIcon from './icons/linkIcon'

interface BlocksContainerProps {
  blocksProp: block[]
  nodeID: number
  successfullNode: boolean
  clickHandler: Function
}

const BlocksContainer: FC<BlocksContainerProps> = (props: {
  blocksProp: block[]
  nodeID: number
  successfullNode: boolean
  clickHandler: Function
}) => {
  const { blocksProp, nodeID, successfullNode, clickHandler } = props
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
              clickHandler,
            }}
          />
          <LinkIcon key={`${block.block.number}-Link`} />
        </div>
      )
    }
    addedBlocks.add(`${nodeID}-${block.block.number}`)
  }
  return <div id={`blocks-${nodeID}`}>{blocks}</div>
}

export default BlocksContainer
