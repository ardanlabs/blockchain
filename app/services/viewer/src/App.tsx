import './App.css'
import React, { Component } from 'react'
import BlocksContainer from './components/blocksContainer'
import TransactionTable from './components/transactionTable'
import Modal from './components/modal'
import nodes from './nodes'
import { transaction, node, block, nodeStatus, mempoolTransaction } from '../types/index.d'
import axios from 'axios';
import MempoolTable from './components/mempoolTable'
import CloseIcon from './components/icons/closeIcon'
import ArrowDownIcon from './components/icons/arrowDownIcon'

// State type is created
export type State = {
  nodes: node[]
  allTransactions: transaction[][][]
  lastBlockHash: string
  blockHashes: Set<string>
  currentNode: node,
  showMempool: boolean,
  mempool: JSX.Element[],
  showTransactions: boolean,
  modalTransactions: JSX.Element[],
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
      mempool: [],
      showTransactions: false,
      modalTransactions: [],
    }
    this.connect = this.connect.bind(this);
    this.reqListener = this.reqListener.bind(this);
    this.handleNewBlock = this.handleNewBlock.bind(this);
    this.changeNodeState = this.changeNodeState.bind(this);
  }
  componentDidMount(): void {
    this.state.nodes.forEach((node) => {
      // If node is inactive it doesn't make the call
      if (node.active) {
        this.connect(node.wsUrl, node.httpUrl, node.nodeID, node.accountID)
      }
    })
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
    ws.onclose = (evt: CloseEvent) => {
      const { connect } = this
      console.log('Socket is closed. Reconnect will be attempted in 1 second.', evt.reason);
      this.changeNodeState('Connecting...', nodeID)
      setTimeout(function() {
        connect(wsUrl, httpUrl, nodeID, accountID);
      }, 1000);
    }
    
    ws.onerror = function(err) {
      console.log(err)
      console.error('Socket encountered error: ', err, 'Closing socket');
      ws.close();
    };
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
    if(node.port) {
      try{
        axios.get(`http://localhost:${node.port}/v1/tx/uncommitted/list`)
        .then(res => {
          const response = res.data;
          console.log(response);
          reqHandler(response);
        })
      } catch (error) {
        console.log(error)
      }
    }
    const reqHandler = (response: mempoolTransaction[]) => {
      if (response.length) {
        const elements: JSX.Element[] = []
        for (let i = 0; i < response.length; i++) {
          const element = response[i]
          const isLast = response.indexOf(element) === response.length - 1
          elements.push(
            <div>
              <MempoolTable key={element.r} transaction={element} />
              <ArrowDownIcon key={`${element.r}-arrow`} isLast={ isLast } />
            </div>
          )
        }
        this.setState({
          mempool: elements,
          currentNode: node,
          showMempool: elements.length ? !this.state.showMempool : false,
        })
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
  blockClickHandler(event: block) {
    const elements: JSX.Element[] = []
    event.trans.forEach(element => {
      const isLast = event.trans.indexOf(element) === event.trans.length - 1
      elements.push(
        <div>
          <TransactionTable key={element.r} transaction={element} />
          <ArrowDownIcon key={`${element.r}-arrow`} isLast={ isLast } />
        </div>
      )
    })
    this.setState({
      modalTransactions: elements,
      showTransactions: !this.state.showTransactions,
    })
  }
  hideTransactionsTable() {
    this.setState({showTransactions: false})
  }
  hideMempoolTable() {
    this.setState({showMempool: false})
  }
  // State is created for the App
  render() {
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
                clickHandler: (evt: block) => this.blockClickHandler(evt)
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
        <Modal classes="transactions" show={this.state.showTransactions}>
          <CloseIcon classes="close-modal-button" onClick={() => this.hideTransactionsTable()} />
          <div id="transactions-content">
            {this.state.modalTransactions}
          </div>
        </Modal>
        <Modal classes="mempool" show={this.state.showMempool}>
          <CloseIcon classes="close-modal-button" onClick={() => this.hideMempoolTable()} />
          <div id="mempool-content">
            {this.state.mempool}
          </div>
        </Modal>
      </div>
    )
  }
}

export default App
