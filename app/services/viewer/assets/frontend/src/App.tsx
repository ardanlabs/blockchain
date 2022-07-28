import './App.css'
import BlockContainer from './components/blockContainer'

function App() {
  const block = {
    block: {
      number: 1,
      prev_block_hash:
        '0x0000000000000000000000000000000000000000000000000000000000000000',
      timestamp: 1651588397,
      beneficiary: '0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8',
      difficulty: 6,
      mining_reward: 700,
      state_root:
        '0x187c4fd4c30c3ae694644dda31978228a9e6326f82384105093e11cb5a0d28a9',
      trans_root:
        '0xf716309d9ece8aa06decc5d21b4062333d8cab4b1b49b7e07110f3f85471c479',
      nonce: 1968008507403245300,
    },
    hash: '0x00000085712967b34f22f848da7d6ce660c7afb2b4f08e1b1252ee890774df11',
    trans: [
      {
        chain_id: 1,
        nonce: 1,
        to: '0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76',
        value: 100,
        tip: 0,
        data: null,
        v: 29,
        r: 9.602438913698971e76,
        s: 1.031947833322543e76,
        timestamp: 1651588397,
        gas_price: 15,
        gas_units: 1,
      },
    ],
  }
  return (
    <div className="App">
      <header className="App-header">
        <h2>Ardan Node Viewer</h2>
      </header>
      <div id="flex-container" className="container-fluid flex-column">
        <div id="msg-block1" className="flex-column">
          <div id="first-msg1" className="block info">
            Node 1: Connecting...
          </div>
          <div className="d-flex" id="blocks1"></div>
          <BlockContainer {...{nodeID: "1", blockNumber: 1, block, successfullNode: true}} />
        </div>
        <div id="msg-block2" className="flex-column">
          <div id="first-msg2" className="block info">
            Node 2: Connecting...
          </div>
          <div className="d-flex" id="blocks2"></div>
        </div>
        <div id="msg-block3" className="flex-column">
          <div id="first-msg3" className="block info">
            Node 3: Connecting...
          </div>
          <div className="d-flex" id="blocks3"></div>
        </div>
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

export default App
