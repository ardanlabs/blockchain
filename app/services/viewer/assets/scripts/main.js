var allTransactions = new Array();

function connect(wsUrl, httpUrl, nodeID, accountID) {
    let blockHashes = new Set();
    let lastBlockHash = "";
    allTransactions.push(new Array());
    
    const handleNewBlock = function(block) {
        let successfullNode = false;
        if (block.hash) {
            if (blockHashes.has(block.hash)) {
                return;
            }
            if (block.block.prev_block_hash === lastBlockHash) {
                addArrow(nodeID);
            }
            blockHashes.add(block.hash);
            lastBlockHash = block.hash;
            allTransactions[nodeID - 1].push(block.trans)
        }
        if (block.block.beneficiary == accountID) {
            successfullNode = true;
        }
        addBlock(nodeID, blockHashes.size, block, successfullNode);
    }

    const reqListener = function() {
        var responseJson = JSON.parse(this.responseText);
        for (i = 0; i < responseJson.length; i++) {
            handleNewBlock(responseJson[i]);
        }
    }

    let socket = new WebSocket(wsUrl);
    socket.onopen = function() {
        document.getElementById(`first-msg${nodeID}`).innerHTML = `Node ${nodeID}: Connection open`;

        var oReq = new XMLHttpRequest();
        oReq.addEventListener("load", reqListener.bind(oReq, nodeID), false);
        oReq.open("GET", httpUrl);
        oReq.send();
    };
  
    socket.onmessage = function(event) {
        let text = event.data;
        if (text.startsWith("viewer: block:")) {
            const blockMsgStart = 'viewer: block:'
            text = text.substring(blockMsgStart.length);
            let block = JSON.parse(text);
            handleNewBlock(block);
            return;
        }
        if (text.includes("MINING: completed")) {
            document.getElementById(`first-msg${nodeID}`).innerHTML = `Node ${nodeID}: Connected`;
            return;
        }
        if (text.includes("MINING: running")) {
            document.getElementById(`first-msg${nodeID}`).innerHTML = `Node ${nodeID}: Mining...`;
            return;
        }
        return;
    };
  
    socket.onclose = function(event) {
        console.log('Socket is closed. Reconnect will be attempted in 1 second.', event.reason);
        document.getElementById(`first-msg${nodeID}`).innerHTML = `Node ${nodeID}: Connecting...`;
        setTimeout(function() {
            connect(wsUrl, httpUrl, nodeID, accountID);
        }, 1000);
    };
  
    socket.onerror = function(err) {
      console.error('Socket encountered error: ', err.message, 'Closing socket');
      socket.close();
    };
}

function getBlockTable(block) {
    if (block.hash) {
        return `
                <table>
                    <tr>
                        <td class="key">Own Hash:</td>
                        <td colspan="5" class="value">${block.hash}</td>
                    </tr>
                    <tr>
                        <td class="key">Previous Hash:</td>
                        <td colspan="5" class="value">${block.block.prev_block_hash}</td>
                    </tr>
                    <tr>
                        <td class="key">Block Number:</td>
                        <td class="value">${block.block.number}</td>
                        <td class="key">Mining Difficulty:</td>
                        <td class="value">${block.block.difficulty}</td>
                        <td class="key">Mining Reward:</td>
                        <td class="value">${block.block.mining_reward}</td>
                    </tr>
                    <tr>
                        <td class="key">Timestamp:</td>
                        <td class="value">${block.block.timestamp}</td>
                        <td class="key">No. of Transactions:</td>
                        <td class="value">${block.trans.length}</td>
                        <td class="key">Nonce:</td>
                        <td class="value">${block.block.nonce}</td>
                    </tr>
                    <tr>
                        <td class="key">Beneficiary:</td>
                        <td colspan="5" class="value">${block.block.beneficiary}</td>
                    </tr>
                    <tr>
                        <td class="key">Transaction Root:</td>
                        <td colspan="5" class="value">${block.block.trans_root}</td>
                    </tr>
                    <tr>
                        <td class="key">State Root:</td>
                        <td colspan="5" class="value">${block.block.state_root}</td>
                    </tr>
                </table>
        `;
    } else {
        return `
            <p>${block.error}</p>
        `;
    }
}

function addBlock(nodeID, blockNumber, block, successfullNode) {
    const blocks = document.getElementById(`blocks${nodeID}`);
    const blockTable = getBlockTable(block);
    console.log(block)
    let extraClass = "";
    if (successfullNode) {
        extraClass = " mine";
    }
    blocks.innerHTML += `
        <div id="block-${nodeID}-${blockNumber}" class="block${extraClass}" onclick="showTransactions(${nodeID}, ${blockNumber})">
            ${blockTable}
        </div>
    `;
    blocks.scrollTop = blocks.scrollHeight;
}

function addArrow(nodeID) {
    const blocks = document.getElementById(`blocks${nodeID}`);
    blocks.innerHTML += `
    <svg style="width:24px;height:24px" viewBox="0 0 24 24">
        <path fill="currentColor" d="M3.9,12C3.9,10.29 5.29,8.9 7,8.9H11V7H7A5,5 0 0,0 2,12A5,5 0 0,0 7,17H11V15.1H7C5.29,15.1 3.9,13.71 3.9,12M8,13H16V11H8V13M17,7H13V8.9H17C18.71,8.9 20.1,10.29 20.1,12C20.1,13.71 18.71,15.1 17,15.1H13V17H17A5,5 0 0,0 22,12A5,5 0 0,0 17,7Z" />
    </svg>
    `;
}

function showTransactions(nodeID, blockNumber) {
    const transactions = document.getElementById("transactions");
    transactions.style.display = "block";

    const transactionsContent = document.getElementById("transactions-content");
    const trans = allTransactions[nodeID - 1][blockNumber - 1];
    for (const t of trans) {
        const transTable = getTransTable(t)
        transactionsContent.innerHTML += `
            <div class="trans">
                ${transTable}
            </div>
        `;
    }
}

function hideTransactions(nodeID, blockNumber) {
    const transactions = document.getElementById("transactions");
    transactions.style.display = "none";

    const transactionsContent = document.getElementById("transactions-content");
    transactionsContent.innerHTML = "";
}

function getTransTable(t) {
    return `
            <table>
                <tr>
                    <td class="key">Chain ID:</td>
                    <td class="value">${t.chain_id}</td>
                    <td class="key">Nonce:</td>
                    <td class="value">${t.nonce}</td>
                    <td class="key">Value:</td>
                    <td class="value">${t.value}</td>
                </tr>
                <tr>
                    <td class="key">To:</td>
                    <td colspan="5" class="value">${t.to}</td>
                </tr>
                <tr>
                    <td class="key">Tip:</td>
                    <td class="value">${t.tip}</td>
                    <td class="key">V:</td>
                    <td class="value">${t.v}</td>
                    <td class="key">Timestamp:</td>
                    <td class="value">${t.timestamp}</td>
                </tr>
                <tr>
                    <td class="key">Gas Price:</td>
                    <td colspan="2" class="value">${t.gas_price}</td>
                    <td class="key">Gas Units:</td>
                    <td colspan="2" class="value">${t.gas_units}</td>
                </tr>
                <tr>
                    <td class="key">Data:</td>
                    <td colspan="5" class="value">${t.data}</td>
                </tr>
                <tr>
                    <td class="key">R:</td>
                    <td colspan="5" class="value">${t.r}</td>
                </tr>
                <tr>
                    <td class="key">S:</td>
                    <td colspan="5" class="value">${t.s}</td>
                </tr>
            </table>
    `;
}

function showMempool(nodeID, port) {
    
    const reqListener = function() {
        const mempool = document.getElementById("mempool");
        mempool.style.display = "block";
    
        const mempoolContent = document.getElementById("mempool-content");
        var responseJson = JSON.parse(this.responseText);
        for (i = 0; i < responseJson.length; i++) {
            mempoolTable = getMempoolTable(responseJson[i]);
            mempoolContent.innerHTML += `
            <div class="trans">
                ${mempoolTable}
            </div>
            `;
        }

    }
    var oReq = new XMLHttpRequest();
    oReq.addEventListener("load", reqListener.bind(oReq, nodeID), false);
    oReq.open("GET", `http://localhost:${port}/v1/tx/uncommitted/list`);
    oReq.send();
}

function hideMempool(nodeID, blockNumber) {
    const transactions = document.getElementById("mempool");
    transactions.style.display = "none";

    const transactionsContent = document.getElementById("mempool-content");
    transactionsContent.innerHTML = "";
}

function getMempoolTable(t) {
    return `
            <table>
                <tr>
                    <td class="key">From:</td>
                    <td colspan="5" class="value">${t.from}</td>
                </tr>
                <tr>
                    <td class="key">From Name:</td>
                    <td colspan="5" class="value">${t.from_name}</td>
                </tr>
                <tr>
                    <td class="key">To:</td>
                    <td colspan="5" class="value">${t.to}</td>
                </tr>
                <tr>
                    <td class="key">To Name:</td>
                    <td colspan="5" class="value">${t.to_name}</td>
                </tr>
                <tr>
                    <td class="key">Chain ID:</td>
                    <td class="value">${t.chain_id}</td>
                    <td class="key">Nonce:</td>
                    <td class="value">${t.nonce}</td>
                    <td class="key">Value:</td>
                    <td class="value">${t.value}</td>
                </tr>
                <tr>
                    <td class="key">Tip:</td>
                    <td class="value">${t.tip}</td>
                    <td class="key">Data:</td>
                    <td class="value">${t.data}</td>
                    <td class="key">Timestamp:</td>
                    <td class="value">${t.timestamp}</td>
                </tr>
                <tr>
                    <td class="key">Gas Price:</td>
                    <td colspan="2" class="value">${t.gas_price}</td>
                    <td class="key">Gas Units:</td>
                    <td colspan="2" class="value">${t.gas_units}</td>
                </tr>
                <tr>
                    <td class="key">Sig:</td>
                    <td colspan="5" class="value">${t.sig}</td>
                </tr>
                <tr>
                    <td class="key">Proof:</td>
                    <td colspan="5" class="value">${t.proof}</td>
                </tr>
                <tr>
                    <td class="key">Proof Order:</td>
                    <td colspan="5" class="value">${t.proof_order}</td>
                </tr>
            </table>
    `;
}

connect('ws://localhost:8080/v1/events', 'http://localhost:9080/v1/node/block/list/1/latest', 1, '0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8');
connect('ws://localhost:8280/v1/events', 'http://localhost:9280/v1/node/block/list/1/latest', 2, '0xb8Ee4c7ac4ca3269fEc242780D7D960bd6272a61');
connect('ws://localhost:8380/v1/events', 'http://localhost:9380/v1/node/block/list/1/latest', 3, '0x616c90073c78ac073D89E750836401a92B16dE7e');