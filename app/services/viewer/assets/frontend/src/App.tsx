import './App.css'
import React, { Component } from 'react'
import MsgBlock from './components/msgBlock'
import nodes from './nodes'
import { transaction, node, block } from '../types/index.d'

// State type is created
export type State = {
  nodes: node[]
  allTransactions: transaction[][]
  lastBlockHash: string
  blockHashes: Set<string>
}

class App extends Component<{}, State> {
  // State is created for the App
  constructor(props: any) {
    super(props)
    this.state = {
      nodes: nodes,
      allTransactions: [],
      lastBlockHash: '',
      blockHashes: new Set(),
    }
  }
  // This function add the blocks once the webhook returns them
  public handleNewBlock(block: block, nodeID: number, accountID: string){
    let successfullNode = false;
    if (block.block.beneficiary === accountID) {
      successfullNode = true;
    }
    // We set the new state to add the blocks and transactions
    this.setState(() => {
      // Here we can push and add
      this.state.nodes[nodeID - 1].blocks.push(block)
      if (!this.state.blockHashes.has(block.hash)) {
        this.state.blockHashes.add(block.hash);
      }
      // But here we need to return an object with the keys to update
      return {
        lastBlockHash: block.hash,
        nodes: [...nodes, {...nodes[nodeID - 1], successfull: successfullNode}]
      }
    })
    console.log(this.state)
  }
  public componentDidMount() {
    let self = this;
    // The connect function triggers the ws connection
    const connect = (
      wsUrl: string,
      httpUrl: string,
      nodeID: number,
      accountID: string,
    ) => {
      // This creates an empty array for each node
      this.state.allTransactions.push([])
      // Request listener handles the entry of new conections
      const reqListener = function(this: any) {
        var responseJson = JSON.parse(this.responseText); 
        for (let i = 0; i < responseJson.length; i++) {
          // for each block received, we mutate the state
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
    // Call to start all nodes, nodes can be found inside './src/nodes.tsx'
    nodes.forEach((node) => {
      // If node is inactive it doesn't make the call
      if (node.active) {
        connect(node.wsUrl, node.httpUrl, node.nodeID, node.accountID)
      }
    })
  }
  render() {
    const msgsBlocks: JSX.Element[] = []
    // We implement the nodes inside the UI grouping them inside an JSX.element array
    nodes.forEach((node) => {
      // If node is inactive it doesn't add it to the UI
      if (node.active) {
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
          { msgsBlocks }
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
