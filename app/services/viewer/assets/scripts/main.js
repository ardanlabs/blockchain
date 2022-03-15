const socket = new WebSocket('ws://localhost:8080/v1/events');

socket.addEventListener('open', function (event) {
    document.getElementById("msg").innerHTML = "connection open";
});

socket.addEventListener('message', function (event) {
    var v = document.getElementById("msg").innerHTML;
    v += v + "<BR />" + event.data;
    document.getElementById("msg").innerHTML = v;
});