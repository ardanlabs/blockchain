const socket = new WebSocket('ws://localhost:8080/v1/events');

socket.addEventListener('open', function (event) {
    document.getElementById("msg").innerHTML = "connection open";
});

socket.addEventListener('message', function (event) {
    document.getElementById("msg").innerHTML += "<BR />" + event.data;
});