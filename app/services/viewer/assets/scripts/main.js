const socket = new WebSocket('ws://localhost:8080/v1/events');

socket.addEventListener('open', function (event) {
    document.getElementById("first-msg").innerHTML = "connection open";
});

socket.addEventListener('message', function (event) {
    msgBlock = document.getElementById("msg-block");
    
    // Create arrow using SVG
    var svg = getNode('svg', { height:32 });

    var line1 = getNode('line', { x1:220, x2:220,  y1:0, y2:32, stroke:'rgb(0,0,0)', 'stroke-opacity':'1.0', 'stroke-width':'2.5'});
    var line2 = getNode('line', { x1:212, x2:220,  y1:24, y2:32, stroke:'rgb(0,0,0)', 'stroke-opacity':'1.0', 'stroke-width':'2.5'});
    var line3 = getNode('line', { x1:228, x2:220,  y1:24, y2:32, stroke:'rgb(0,0,0)', 'stroke-opacity':'1.0', 'stroke-width':'2.5'});
    svg.appendChild(line1);
    svg.appendChild(line2);
    svg.appendChild(line3);
    msgBlock.appendChild(svg);

    // Create Text block
    const newMsg = document.createElement("div");
    newMsg.setAttribute("class", "block-class");
    const node = document.createTextNode(event.data);
    newMsg.appendChild(node);
    msgBlock.appendChild(newMsg);
});

function getNode(n, v) {
    n = document.createElementNS("http://www.w3.org/2000/svg", n);
    for (var p in v)
        n.setAttributeNS(null, p, v[p]);
    return n
}