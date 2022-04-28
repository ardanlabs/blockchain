const socket1 = new WebSocket('ws://localhost:8080/v1/events');

socket1.addEventListener('open', function (event) {
    document.getElementById("first-msg1").innerHTML = "connection open";
});

socket1.addEventListener('message', function (event) {
	text = event.data;
    msgBlock = document.getElementById("msg-block1");
    msgBlock.innerHTML += `
	    <svg version="1.1" xmlns="http://www.w3.org/2000/svg" height="32" viewBox="0 0 300 34">
	        <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="220" y1="0" x2="220" y2="32"/>
	        <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="212" y1="24" x2="220" y2="32"/>
	        <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="228" y1="24" x2="220" y2="32"/>
	    </svg>
		<div class="block-class">${text}</div>
`;
});

/*
 * Do all the same for the second socket/node.
 */
const socket2 = new WebSocket('ws://localhost:8180/v1/events');

socket1.addEventListener('open', function (event) {
    document.getElementById("first-msg2").innerHTML = "connection open";
});

socket2.addEventListener('message', function (event) {
	text = event.data;
    msgBlock = document.getElementById("msg-block2");
    msgBlock.innerHTML += `
	    <svg version="1.1" xmlns="http://www.w3.org/2000/svg" height="32" viewBox="0 0 300 34">
	        <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="220" y1="0" x2="220" y2="32"/>
	        <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="212" y1="24" x2="220" y2="32"/>
	        <line stroke="rgb(0,0,0)" stroke-opacity="1.0" stroke-width="2.5" x1="228" y1="24" x2="220" y2="32"/>
	    </svg>
		<div class="block-class">${text}</div>
`;
});
