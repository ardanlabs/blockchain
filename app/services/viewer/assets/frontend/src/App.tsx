import './App.css'
import React, { Component } from 'react'
import MsgBlock from './components/msgBlock'
import nodes from './nodes'
import { transaction, node, block } from '../types/index.d'

// State type is created
export type State = {
  nodes: node[]
  allTransactions: transaction[][][]
  lastBlockHash: string
  blockHashes: Set<string>
}

class App extends Component<{}, State> {
  constructor(props: any) {
    super(props)
    this.state = {
      nodes: nodes,
      allTransactions: [],
      lastBlockHash: '',
      blockHashes: new Set(),
    }
    this.connect = this.connect.bind(this);
    this.handleNewBlock = this.handleNewBlock.bind(this);
  }
  // The connect function triggers the ws connection
  connect(
    wsUrl: string,
    httpUrl: string,
    nodeID: number,
    accountID: string,
  ){
    const self = this
    // This creates an empty array for each node
    this.state.allTransactions.push([])
    // Request listener handles the entry of new conections
    const reqListener = function (this: any) {
      var responseJson = JSON.parse(this.responseText);
      for (let i = 0; i < responseJson.length; i++) {
        // for each block received, we mutate the state
        console.log(responseJson[i], 'response')
        self.handleNewBlock(responseJson[i], nodeID, accountID);
      }
    }

    const ws = new WebSocket(wsUrl)
    ws.onopen = () => {
      var oReq = new XMLHttpRequest();
      // we wait for the request to load and then call the reqListener
      // also nodeID and the request itself are binded to the function to be used inside with 'this'
      oReq.addEventListener('load', reqListener.bind(oReq, nodeID), false)
      oReq.open("GET", httpUrl)
      oReq.send()
    }
    ws.onmessage = (evt: MessageEvent) => {
      const data: any = JSON.parse(evt.data)
      console.log(data)
      // OnMessage functionality to be migrated, still working on it
    }
  }
  handleNewBlock(block: block, nodeID: number, accountID: string) {
    let successfullNode = false;
    if (block.hash) {
      if (this.state.blockHashes.has(block.hash)) {
        return;
      }
      
      this.setState((prevState) => {
        const modifiedAllTransactions = prevState.allTransactions
        modifiedAllTransactions[nodeID - 1].push([...block.trans])
        const modifiedBlockHashes = prevState.blockHashes
        modifiedBlockHashes.add(block.hash)
        return {
          allTransactions: modifiedAllTransactions,
          blockHashes: modifiedBlockHashes,
          lastBlockHash: block.hash,
        }
      })
    }
    if (block.block.beneficiary === accountID) {
      successfullNode = true;
      console.log(successfullNode)
    }
    this.setState((prevState) => {
      const modifiedNodes = prevState.nodes
      
      modifiedNodes[nodeID - 1].successfull = successfullNode
      modifiedNodes[nodeID - 1].blocks.push(block)
      return {
        nodes: modifiedNodes
      }
    })
  }
  // State is created for the App
  render() {
    this.state.nodes.forEach((node) => {
      // If node is inactive it doesn't make the call
      if (node.active) {
        this.connect(node.wsUrl, node.httpUrl, node.nodeID, node.accountID)
      }
    })
    const msgsBlocks: JSX.Element[] = []
    // We implement the nodes inside the UI grouping them inside an JSX.element array
    nodes.forEach((node) => {
      // If node is inactive it doesn't add it to the UI
      if (node.active) {
        console.log(node.successfull)
        msgsBlocks.push(
          <MsgBlock
            key={node.nodeID}
            // MsgBlock props
            {...{
              nodeID: node.nodeID,
              nodeState: node.state,
              blocksProp: node.blocks,
              successfullNode: node.successfull,
            }}
          />,
        )
      }
    })
    return (
      <div className="App">
        <header className="App-header">
          <h2>Ardan Node Viewer</h2>
        </header>
        <div id="flex-container" className="container-fluid flex-column">
          {msgsBlocks}
        </div>
        <div id="transactions">
          <button className="button-transactions-hide">
            <strong>X</strong>
          </button>
          <div id="transactions-content"></div>
        </div>
        <div id="mempool">
          <button className="button-transactions-hide">
            <strong>X</strong>
          </button>
          <div id="mempool-content"></div>
        </div>
      </div>
    )
  }
}

export default App
