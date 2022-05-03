function reqListener(id, evt) {
	console.log(this.responseText);
    var responseJson = JSON.parse(this.responseText);
    msgBlock = document.getElementById(`msg-block${id}`);
    for(i = 0; i < responseJson.length; i++) {
        msgBlock.innerHTML += `
            <svg version="1.1" xmlns="http://www.w3.org/2000/svg" height="32" viewBox="0 0 300 34">
                <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="220" y1="0" x2="220" y2="32"/>
                <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="212" y1="24" x2="220" y2="32"/>
                <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="228" y1="24" x2="220" y2="32"/>
            </svg>
            <div class="block-class">
                <table>
                    <tr>
                        <td class="key">Own Hash:</td>
                        <td colspan="5" class="value">${responseJson[i].hash}</td>
                    </tr>
                    <tr>
                        <td class="key">Previous Hash:</td>
                        <td colspan="5" class="value">${responseJson[i].block.prev_block_hash}</td>
                    </tr>
                    <tr>
                        <td class="key">Block Number:</td>
                        <td class="value">${responseJson[i].block.number}</td>
                        <td class="key">Mining Difficulty:</td>
                        <td class="value">${responseJson[i].block.difficulty}</td>
                        <td class="key">Mining Reward:</td>
                        <td class="value">${responseJson[i].block.mining_reward}</td>
                    </tr>
                    <tr>
                        <td class="key">Timestamp:</td>
                        <td colspan="2" class="value">${responseJson[i].block.timestamp}</td>
                        <td class="key">Nonce:</td>
                        <td colspan="2" class="value">${responseJson[i].block.nonce}</td>
                    </tr>
                    <tr>
                        <td class="key">Beneficiary:</td>
                        <td colspan="5" class="value">${responseJson[i].block.beneficiary}</td>
                    </tr>
                    <tr>
                        <td class="key">Transaction Root:</td>
                        <td colspan="5" class="value">${responseJson[i].block.trans_root}</td>
                    </tr>
                    <tr>
                        <td class="key">State Root:</td>
                        <td colspan="5" class="value">${responseJson[i].block.state_root}</td>
                    </tr>
                </table>
            </div>
        `;
        msgBlock.scrollTop = msgBlock.scrollHeight;
    }
}

function connect(wsUrl, httpUrl, id) {
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
        msgBlock = document.getElementById(`msg-block${id}`);
        msgBlock.innerHTML += `
            <svg version="1.1" xmlns="http://www.w3.org/2000/svg" height="32" viewBox="0 0 300 34">
                <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="220" y1="0" x2="220" y2="32"/>
                <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="212" y1="24" x2="220" y2="32"/>
                <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="228" y1="24" x2="220" y2="32"/>
            </svg>
            <div class="block-class">
                <table>
                    <tr>
                        <td class="key">Own Hash:</td>
                        <td colspan="5" class="value">${block.hash}</td>
                    </tr>
                    <tr>
                        <td class="key">Previous Hash:</td>
                        <td colspan="5" class="value">${block.header.prev_block_hash}</td>
                    </tr>
                    <tr>
                        <td class="key">Block Number:</td>
                        <td class="value">${block.header.number}</td>
                        <td class="key">Mining Difficulty:</td>
                        <td class="value">${block.header.difficulty}</td>
                        <td class="key">Mining Reward:</td>
                        <td class="value">${block.header.mining_reward}</td>
                    </tr>
                    <tr>
                        <td class="key">Timestamp:</td>
                        <td colspan="2" class="value">${block.header.timestamp}</td>
                        <td class="key">Nonce:</td>
                        <td colspan="2" class="value">${block.header.nonce}</td>
                    </tr>
                    <tr>
                        <td class="key">Beneficiary:</td>
                        <td colspan="5" class="value">${block.header.beneficiary}</td>
                    </tr>
                    <tr>
                        <td class="key">Transaction Root:</td>
                        <td colspan="5" class="value">${block.header.trans_root}</td>
                    </tr>
                    <tr>
                        <td class="key">State Root:</td>
                        <td colspan="5" class="value">${block.header.state_root}</td>
                    </tr>
                </table>
            </div>
        `;
        msgBlock.scrollTop = msgBlock.scrollHeight;
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

connect('ws://localhost:8080/v1/events', 'http://localhost:9080/v1/node/block/list/1/latest', '1');
connect('ws://localhost:8180/v1/events', 'http://localhost:9180/v1/node/block/list/1/latest', '2');
connect('ws://localhost:8280/v1/events', 'http://localhost:9280/v1/node/block/list/1/latest', '3');
