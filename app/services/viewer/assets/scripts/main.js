function connect(url, id) {
    let socket = new WebSocket(url);
    socket.onopen = function() {
      document.getElementById(`first-msg${id}`).innerHTML = `Node ${id}: Connection open`;
      msgBlock = document.getElementById(`msg-block${id}`);
        msgBlock.innerHTML += `
            <svg version="1.1" xmlns="http://www.w3.org/2000/svg" height="32" viewBox="0 0 300 34">
                <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="220" y1="0" x2="220" y2="32"/>
                <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="212" y1="24" x2="220" y2="32"/>
                <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="228" y1="24" x2="220" y2="32"/>
            </svg>
            <div id="log-block${id}" class="log-block-class"></div>
        `;
    };
  
    socket.onmessage = function(event) {
        text = event.data;
        msgBlock = document.getElementById(`log-block${id}`);
        msgBlock.innerHTML += "<BR />" + text;
        msgBlock.scrollTop = msgBlock.scrollHeight;
    };
  
    socket.onclose = function(event) {
        console.log('Socket is closed. Reconnect will be attempted in 1 second.', event.reason);
        document.getElementById(`first-msg${id}`).innerHTML = `Node ${id}: Connecting...`;
        setTimeout(function() {
            connect(url, id);
        }, 1000);
    };
  
    socket.onerror = function(err) {
      console.error('Socket encountered error: ', err.message, 'Closing socket');
      socket.close();
    };
  }
  
connect('ws://localhost:8080/v1/events', '1');
connect('ws://localhost:8180/v1/events', '2');
connect('ws://localhost:8280/v1/events', '3');
