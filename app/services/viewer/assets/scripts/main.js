var allTransactions = new Array();

function connect(wsUrl, httpUrl, nodeID) {
    let blockHashes = new Set();
    let lastBlockHash = "";
    let successfullNode = false;
    allTransactions.push(new Array());

    const handleNewBlock = function(block) {
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
        addBlock(nodeID, blockHashes.size, block, successfullNode);
        successfullNode = false;
    }

    const reqListener = function() {
        var responseJson = JSON.parse(this.responseText);
        msgBlock = document.getElementById(`msg-block${nodeID}`);
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
        const blockMsgStart = 'viewer: block: ';
        let text = event.data;
        if (text.startsWith(blockMsgStart)) {
            text = text.substring(blockMsgStart.length);
            let block = JSON.parse(text);
            handleNewBlock(block);
            return;
        }
        if (text.includes("MINING: completed")) {
            document.getElementById(`first-msg${nodeID}`).innerHTML = `Node ${nodeID}: Connected`;
            return;
        }

        if (text.includes("MINING: SOLVED")) {
            successfullNode = true;
            return;
        }
        if (text.includes("MINING")) {
            document.getElementById(`first-msg${nodeID}`).innerHTML = `Node ${nodeID}: Mining...`;
            return;
        }
        return;
    };
  
    socket.onclose = function(event) {
        console.log('Socket is closed. Reconnect will be attempted in 1 second.', event.reason);
        document.getElementById(`first-msg${nodeID}`).innerHTML = `Node ${nodeID}: Connecting...`;
        setTimeout(function() {
            connect(wsUrl, httpUrl, nodeID);
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
    const msgBlock = document.getElementById(`msg-block${nodeID}`);
    const blockTable = getBlockTable(block);
    let extraClass = "";
    if (successfullNode) {
        extraClass = " mine";
    }
    msgBlock.innerHTML += `
        <div id="block-${nodeID}-${blockNumber}" class="block${extraClass}" onclick="showTransactions(${nodeID}, ${blockNumber})">
            ${blockTable}
        </div>
    `;
    msgBlock.scrollTop = msgBlock.scrollHeight;
}

function addArrow(nodeID) {
    const msgBlock = document.getElementById(`msg-block${nodeID}`);
    msgBlock.innerHTML += `
        <svg version="1.1" xmlns="http://www.w3.org/2000/svg" height="32" viewBox="0 0 300 34">
            <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="220" y1="0" x2="220" y2="32"/>
            <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="212" y1="24" x2="220" y2="32"/>
            <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="228" y1="24" x2="220" y2="32"/>
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

connect('ws://localhost:8080/v1/events', 'http://localhost:9080/v1/node/block/list/1/latest', 1);
connect('ws://localhost:8280/v1/events', 'http://localhost:9280/v1/node/block/list/1/latest', 2);
connect('ws://localhost:8380/v1/events', 'http://localhost:9380/v1/node/block/list/1/latest', 3);
