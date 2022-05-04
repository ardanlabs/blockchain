function connect(wsUrl, httpUrl, id) {
    let blockHashes = new Set();
    let lastBlockHash = "";

    const handleNewBlock = function(block) {
        if (blockHashes.size === 0) {
            document.getElementById(`first-msg${id}`).innerHTML = getBlockTable(block);
            if (block.hash) {
                blockHashes.add(block.hash);
                lastBlockHash = block.hash;
            }
            return;
        }
        if (block.hash) {
            if (blockHashes.has(block.hash)) {
                return;
            }
            if (block.block.prev_block_hash === lastBlockHash) {
                addArrow(id);
            }
            blockHashes.add(block.hash);
            lastBlockHash = block.hash;
        }
        addBlock(id, block);
    }

    const reqListener = function() {
        var responseJson = JSON.parse(this.responseText);
        msgBlock = document.getElementById(`msg-block${id}`);
        for (i = 0; i < responseJson.length; i++) {
            handleNewBlock(responseJson[i]);
        }
    }

    let socket = new WebSocket(wsUrl);
    socket.onopen = function() {
        document.getElementById(`first-msg${id}`).innerHTML = `Node ${id}: Connection open`;

        var oReq = new XMLHttpRequest();
        oReq.addEventListener("load", reqListener.bind(oReq, id), false);
        oReq.open("GET", httpUrl);
        oReq.send();
    };
  
    socket.onmessage = function(event) {
        const blockMsgStart = 'viewer: block: ';
        let text = event.data;
        if (!text.startsWith(blockMsgStart)) {
            return;
        }
        text = text.substring(blockMsgStart.length);
        let block = JSON.parse(text);
        handleNewBlock(block);
    };
  
    socket.onclose = function(event) {
        console.log('Socket is closed. Reconnect will be attempted in 1 second.', event.reason);
        document.getElementById(`first-msg${id}`).innerHTML = `Node ${id}: Connecting...`;
        setTimeout(function() {
            connect(wsUrl, httpUrl, id);
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

function addBlock(id, block) {
    msgBlock = document.getElementById(`msg-block${id}`);
    blockTable = getBlockTable(block)
    msgBlock.innerHTML += `
        <div class="block-class">
            ${blockTable}
        </div>
    `;
    msgBlock.scrollTop = msgBlock.scrollHeight;
}

function addArrow(id) {
    msgBlock = document.getElementById(`msg-block${id}`);
    msgBlock.innerHTML += `
        <svg version="1.1" xmlns="http://www.w3.org/2000/svg" height="32" viewBox="0 0 300 34">
            <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="220" y1="0" x2="220" y2="32"/>
            <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="212" y1="24" x2="220" y2="32"/>
            <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="228" y1="24" x2="220" y2="32"/>
        </svg>
    `;
}

connect('ws://localhost:8080/v1/events', 'http://localhost:9080/v1/node/block/list/1/latest', '1');
connect('ws://localhost:8280/v1/events', 'http://localhost:9280/v1/node/block/list/1/latest', '2');
connect('ws://localhost:8380/v1/events', 'http://localhost:9380/v1/node/block/list/1/latest', '3');
