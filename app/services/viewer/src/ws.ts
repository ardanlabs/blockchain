import nodes from './nodes'
import { transaction, node, block } from '../types/index.d'
// constructor(props: any) {
//   super(props)
//   this.state = {
//     nodes: nodes,
//     allTransactions: [],
//     lastBlockHash: '',
//     blockHashes: new Set(),
//   }
// }
// The connect function triggers the ws connection
// export const connect2 = (
//   wsUrl: string,
//   httpUrl: string,
//   nodeID: number,
//   accountID: string,
// ) => {

// const handleNewBlock = function (block) {
//   let successfullNode = false;
//   if (block.hash) {
//     if (blockHashes.has(block.hash)) {
//       return;
//     }
//     if (block.block.prev_block_hash === lastBlockHash) {
//       addArrow(nodeID);
//     }
//     blockHashes.add(block.hash);
//     lastBlockHash = block.hash;
//     allTransactions[nodeID - 1].push(block.trans)
//   }
//   if (block.block.beneficiary == accountID) {
//     successfullNode = true;
//   }
//   console.log(allTransactions, 'all')
//   addBlock(nodeID, blockHashes.size, block, successfullNode);
// }
//   // This creates an empty array for each node
//   this.state.allTransactions.push([])
//   // Request listener handles the entry of new conections
//   const reqListener = function (this: any) {
//     var responseJson = JSON.parse(this.responseText);
//     for (let i = 0; i < responseJson.length; i++) {
//       // for each block received, we mutate the state
//       self.handleNewBlock(responseJson[i], nodeID, accountID);
//     }
//   }

//   const ws = new WebSocket(wsUrl)
//   ws.onopen = () => {
//     var oReq = new XMLHttpRequest();
//     // we wait for the request to load and then call the reqListener
//     // also nodeID and the request itself are binded to the function to be used inside with 'this'
//     oReq.addEventListener('load', reqListener.bind(oReq, nodeID), false)
//     oReq.open("GET", httpUrl)
//     oReq.send()
//   }
//   ws.onmessage = (evt: MessageEvent) => {
//     const data: any = JSON.parse(evt.data)
//     console.log(data)
//     // OnMessage functionality to be migrated, still working on it
//   }
// }
let allTransactions = new Array();
export function connect(
  wsUrl: string,
  httpUrl: string,
  nodeID: number,
  accountID: string,
) {
  let blockHashes = new Set();
  let lastBlockHash = "";
  allTransactions.push(new Array());
  
  const handleNewBlock = function(block: block) {
      let successfullNode = false;
      if (block.hash) {
          if (blockHashes.has(block.hash)) {
              return;
          }
          blockHashes.add(block.hash);
          lastBlockHash = block.hash;
          allTransactions[nodeID - 1].push(block.trans)
      }
      if (block.block.beneficiary == accountID) {
          successfullNode = true;
      }

      nodes[nodeID - 1].blocks.push(block)
  }

  const reqListener = function(this: any) {
    var responseJson = JSON.parse(this.responseText);
    for (let i = 0; i < responseJson.length; i++) {
        handleNewBlock(responseJson[i]);
    }
  }

  let socket = new WebSocket(wsUrl);
  socket.onopen = function() {
      nodes[nodeID - 1].state = `Connection open`;

      var oReq = new XMLHttpRequest();
      oReq.addEventListener("load", reqListener.bind(oReq, nodeID), false);
      oReq.open("GET", httpUrl);
      oReq.send();
  };

  socket.onmessage = function(event) {
      let text = event.data;
      console.log(event)
      if (text.startsWith("viewer: block:")) {
          const blockMsgStart = 'viewer: block:'
          text = text.substring(blockMsgStart.length);
          let block = JSON.parse(text);
          handleNewBlock(block);
          return;
      }
      if (text.includes("MINING: completed")) {
          nodes[nodeID - 1].state = `Connected`;
          return;
      }
      if (text.includes("MINING: running")) {
          nodes[nodeID - 1].state = `Mining...`;
          return;
      }
      return;
  };

  socket.onclose = function(event) {
      console.log('Socket is closed. Reconnect will be attempted in 1 second.', event.reason);
      nodes[nodeID - 1].state = `Connecting...`;
      setTimeout(function() {
          connect(wsUrl, httpUrl, nodeID, accountID);
      }, 1000);
  };

  socket.onerror = function(error) {
    console.error('Socket encountered error: ', error, 'Closing socket');
    socket.close();
  };
}