import './App.css'
import React, { Component } from 'react'
import BlocksContainer from './components/blocksContainer'
import Modal from './components/modal'
import nodes from './nodes'
import { transaction, node, block, nodeStatus } from '../types/index.d'
import axios from 'axios';

// State type is created
export type State = {
  nodes: node[]
  allTransactions: transaction[][][]
  lastBlockHash: string
  blockHashes: Set<string>
  currentNode: node,
  showMempool: boolean,
  showTransactions: boolean,
}

interface mempoolResponse {
  data: object[]
}

class App extends Component<{}, State> {
  constructor(props: any) {
    super(props)
    this.state = {
      nodes: [...nodes],
      allTransactions: [],
      lastBlockHash: '',
      blockHashes: new Set(),
      currentNode: {} as node,
      showMempool: false,
      showTransactions: false,
    }
    this.connect = this.connect.bind(this);
    this.reqListener = this.reqListener.bind(this);
    this.handleNewBlock = this.handleNewBlock.bind(this);
    this.changeNodeState = this.changeNodeState.bind(this);
  }
  // The connect function triggers the ws connection
  reqListener(response: block[], nodeID: number, accountID: string) {
    if(response.length) {
      for (let i = 0; i < response.length; i++) {
        // for each block received, we mutate the state
        this.handleNewBlock(response[i], nodeID, accountID)
      }
    }
  }
  componentDidMount(): void {
    this.state.nodes.forEach((node) => {
      // If node is inactive it doesn't make the call
      if (node.active) {
        this.connect(node.wsUrl, node.httpUrl, node.nodeID, node.accountID)
      }
    })
  }
  connect(
    wsUrl: string,
    httpUrl: string,
    nodeID: number,
    accountID: string,
  ){
    // This creates an empty array for each node
    this.state.allTransactions.push([])
    // Request listener handles the entry of new conections

    const ws = new WebSocket(wsUrl)
    ws.onopen = () => {
      this.changeNodeState('Connection open', nodeID)
      try{
        axios.get(httpUrl)
        .then(res => {
          const response = res.data;      
          this.reqListener(response, nodeID, accountID);
        })
      } catch (error) {
        console.error(error)
      }
    }
    ws.onmessage = (evt: MessageEvent) => {
      if (evt.data) {
        let text = evt.data        
        if (text.startsWith("viewer: block:")) {
            const blockMsgStart = 'viewer: block:'
            text = text.substring(blockMsgStart.length);
            let block = JSON.parse(text);
            this.handleNewBlock(block, nodeID, accountID);
            return;
        }
        if (text.includes("MINING: completed")) {
            this.changeNodeState('Connected', nodeID)
            return;
        }
        if (text.includes("MINING: running")) {
          console.info(text.replace('viewer: ', ''))
          this.changeNodeState('Mining...', nodeID)
            return;
        }
      }
      return;
    }
  }
  changeNodeState(status: nodeStatus, nodeID: number){
    this.setState((prevState) => {
      const modifiedNodes = prevState.nodes
      modifiedNodes[nodeID - 1].state = status
      return {
        nodes: modifiedNodes
      }
    })
  }
  showMempool(node: node){
    this.setState({
      currentNode: node,
      showMempool: !this.state.showMempool,
    })
    console.log(!this.state.showMempool)
    if(node.port) {
      try{
        axios.get(`http://localhost:${node.port}/v1/tx/uncommitted/list`)
        .then(res => {
          const response = res.data;
          reqHandler(response);
        })
      } catch (error) {
        console.log(error)
      }
    }
    const reqHandler = (response: mempoolResponse) => {
      if ('data' in response && response.data.length) {
        for (let i = 0; i < response.data.length; i++) {
          console.log(response.data[i])
        }
      }
    }
  }
  handleNewBlock(block: block, nodeID: number, accountID: string) {
    let successfullNode = false;
    if (block.hash) {
      if (this.state.blockHashes.has(`${accountID}${nodeID}${block.hash}`)) {
        return;
      }
      
      this.setState((prevState) => {
        const modifiedAllTransactions = prevState.allTransactions
        modifiedAllTransactions[nodeID - 1].push([...block.trans])
        const modifiedBlockHashes = prevState.blockHashes
        modifiedBlockHashes.add(`${accountID}${nodeID}${block.hash}`)
        return {
          allTransactions: modifiedAllTransactions,
          blockHashes: modifiedBlockHashes,
          lastBlockHash: block.hash,
        }
      })
    }
    if (block.block.beneficiary === accountID) {
      successfullNode = true;
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
  componentDidUpdate(prevProps: Readonly<{}>, prevState: Readonly<State>, snapshot?: any): void {
    console.log(prevState)
  }
  // State is created for the App
  render() {
    console.log(this.state.showTransactions, 'trans show')
    const msgsBlocks: JSX.Element[] = []
    this.state.nodes.forEach((node) => {
      // We implement the nodes inside the UI grouping them inside an JSX.element array
      const { nodeID, state, blocks, successfull } = node
      // If node is inactive it doesn't add it to the UI
      if (node.active) {
        msgsBlocks.push(
          <div key={nodeID + 'msg-block'} id={`msg-block${nodeID}`} className="flex-column">
            <div id={`first-msg${nodeID}`} className="block info" onClick={() => this.showMempool(node)}>
              Node {nodeID}: {state}
            </div>
            <BlocksContainer
              key={nodeID}
              {...{
                nodeID: nodeID,
                blocksProp: blocks,
                successfullNode: successfull,
              }}
            />
        </div>
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
        <Modal classes="transactions" show={this.state.showMempool}>
          <button className="button-transactions-hide">
            <strong>X</strong>
          </button>
          <div id="transactions-content"></div>
        </Modal>
        <Modal classes="mempool" show={this.state.showMempool}>
          <button className="button-transactions-hide" onClick={() => this.showMempool({} as node)}>
            <strong>X</strong>
          </button>
          <div className="trans">Node {this.state.currentNode.nodeID}</div>
          <div id="mempool-content"></div>
        </Modal>
      </div>
    )
  }
}

export default App
