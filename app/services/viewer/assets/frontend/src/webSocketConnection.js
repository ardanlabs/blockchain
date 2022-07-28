import nodes from './nodes'
var allTransactions = []
export default function connect(wsUrl, httpUrl, nodeID, accountID) {
  let blockHashes = new Set()
  // let lastBlockHash = ''
  allTransactions.push([])
  const handleNewBlock = function (block) {
    let successfullNode = false
    if (block.hash) {
      if (blockHashes.has(block.hash)) {
        return
      }
      blockHashes.add(block.hash)
      // lastBlockHash = block.hash
      allTransactions[nodeID - 1].push(block.trans)
    }
    if (block.block.beneficiary === accountID) {
      successfullNode = true
    }
   console.log(nodeID, blockHashes.size, block, successfullNode)
  }
  const reqListener = function () {
    var responseJson = JSON.parse(this.responseText)
    for (let i = 0; i < responseJson.length; i++) {
      handleNewBlock(responseJson[i])
    }
  }
  let socket = new WebSocket(wsUrl)
  socket.onopen = function () {
    nodes[nodeID].state = 'Connection open'
    var oReq = new XMLHttpRequest()
    oReq.addEventListener('load', reqListener.bind(oReq, nodeID), false)
    oReq.open('GET', httpUrl)
    oReq.send()
  }
  socket.onmessage = function (event) {
    let text = event.data
    if (text.startsWith('viewer: block:')) {
      const blockMsgStart = 'viewer: block:'
      text = text.substring(blockMsgStart.length)
      let block = JSON.parse(text)
      handleNewBlock(block)
      return
    }
    if (text.includes('MINING: completed')) {
      nodes[nodeID].state = 'Connected'
      return
    }
    if (text.includes('MINING: running')) {
      nodes[nodeID].state = 'Mining...'
      return
    }
    return
  }
  socket.onclose = function (event) {
    console.log(
      'Socket is closed. Reconnect will be attempted in 1 second.',
      event.reason,
    )
    nodes[nodeID].state = 'Connecting...'
    setTimeout(function () {
      connect(wsUrl, httpUrl, nodeID, accountID)
    }, 1000)
  }
  socket.onerror = function (err) {
    console.error('Socket encountered error: ', err.message, 'Closing socket')
    socket.close()
  }
}
