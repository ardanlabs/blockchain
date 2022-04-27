const socket = new WebSocket('ws://localhost:8080/v1/events');

socket.addEventListener('open', function (event) {
    document.getElementById("first-msg").innerHTML = "connection open";
});

socket.addEventListener('message', function (event) {
    msgBlock = document.getElementById("msg-block");
	text = event.data;
    msgBlock.innerHTML += `
	    <svg version="1.1" xmlns="http://www.w3.org/2000/svg" height="32" viewBox="0 0 300 34">
	        <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="220" y1="0" x2="220" y2="32"/>
	        <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="212" y1="24" x2="220" y2="32"/>
	        <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="228" y1="24" x2="220" y2="32"/>
	    </svg>
		<div class="block-class">${text}</div>
`;
});
