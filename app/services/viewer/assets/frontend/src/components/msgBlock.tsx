import React, { FC } from 'react'
import { block } from '../../types/index.d'
import BlocksContainer from './blocksContainer'

interface MsgBlockProps {
  blocksProp: block[],
  nodeID: number,
  nodeState: string,
  successfullNode: boolean,
}
const MsgBlock: FC<MsgBlockProps> = ({
  blocksProp,
  nodeID,
  nodeState,
  successfullNode,
}) => {
  return (
    <div id="msg-block1" className="flex-column">
      <div id="first-msg1" className="block info">
        Node {nodeID}: {nodeState}
      </div>
      <BlocksContainer
        {...{
          nodeID: nodeID,
          blockNumber: 1,
          blocksProp: blocksProp,
          successfullNode: successfullNode,
        }}
      />
    </div>
  )
}

export default MsgBlock
