$.ajaxSetup({
    contentType: "application/json; charset=utf-8",
    error: function (xhr) {
        const conn = document.getElementById("connected");
        conn.className = "notconnected";
        conn.innerHTML = "NOT CONNECTED";
      }
});

window.onload = function () {
    wireEvents();
    showInfoTab("send");
    connect();
}

function wireEvents() {
    const from = document.getElementById("from");
    from.addEventListener(
        'change',
        fromBalance,
        false
    );

    const to = document.getElementById("to");
    to.addEventListener(
        'change',
        toBalance,
        false
    );

    const send = document.getElementById("sendbutton");
    send.addEventListener(
        'click',
        showInfoTabSend,
        false
    );

    const tran = document.getElementById("tranbutton");
    tran.addEventListener(
        'click',
        showInfoTabTran,
        false
    );

    const sendsubmit = document.getElementById("sendsubmit");
    sendsubmit.addEventListener(
        'click',
        submitTran,
        false
    );

    const sendamount = document.getElementById("sendamount");
    sendamount.addEventListener(
        'keyup',
        formatCurrencyKeyup,
        false
    );
    sendamount.addEventListener(
        'blur',
        formatCurrencyBlur,
        false
    );
}

// =============================================================================

function connect() {
    const url = "http://localhost:8080/v1/genesis/list"

    $.get(url, function (o, status) {
        const conn = document.getElementById("connected");

        if ((typeof o.errors != "undefined") && (o.errors.length > 0)) {    
            conn.className = "notconnected";
            conn.innerHTML = "NOT CONNECTED";
            return;
        }

        conn.className = "connected";
        conn.innerHTML = "CONNECTED";

        fromBalance();
        toBalance();

        return
    });
}

// =============================================================================

function fromBalance() {
    var wallet = new ethers.Wallet(document.getElementById("from").value);
    const url = "http://localhost:8080/v1/accounts/list/" + wallet.address;

    $.get(url, function (o, status) {
        if ((typeof o.errors != "undefined") && (o.errors.length > 0)) {    
            window.alert("ERROR: " + o.errors[0].message);
            return;
        }

        const bal = document.getElementById("frombal");
        bal.innerHTML = formatter.format(o.accounts[0].balance) + " ARD";
    });
}

function toBalance() {
    const url = "http://localhost:8080/v1/accounts/list/" + document.getElementById("to").value;

    $.get(url, function (o, status) {
        if ((typeof o.errors != "undefined") && (o.errors.length > 0)) {    
            window.alert("ERROR: " + o.errors[0].message);
            return;
        }

        const bal = document.getElementById("tobal");
        bal.innerHTML = formatter.format(o.accounts[0].balance) + " ARD";
    });
}

// =============================================================================

function submitTran() {

    // Capture the amount to send.
    const amount = document.getElementById("sendamount").value.replace(/\$|,/g, '');
    
    // Construct a userTx with all the information.
    const userTx = {
        nonce: 11,
        to: document.getElementById("to").value,
        value: Number(amount),
        tip: 10,
        data: null,
    };

    // Marshal the userTx to a string and convert the string to bytes.
    const marshal = JSON.stringify(userTx);
    const marshalBytes = ethers.utils.toUtf8Bytes(marshal);

    // Hash the transaction data into a 32 byte array. This will provide
	// a data length consistency with all transactions.
    const txHash = ethers.utils.keccak256(marshalBytes);
    const bytes = ethers.utils.arrayify(txHash);

    // Now sign the data. The underlying code will apply the Ardan stamp and
    // ID to the signature thanks to changes made to the ether.js api.
    const wallet = new ethers.Wallet(document.getElementById("from").value);
    signature = wallet.signMessage(bytes);

    // Since everything is built on promises, wait for the signature to
    // be calculated and then send the transaction to the node.
    signature.then((sig) => sendTran(userTx, sig));
}

function sendTran(userTx, sig) {

    // Need to break out the R and S bytes from the signature.
    const byt = ethers.utils.arrayify(sig);
    const rSlice = byt.slice(0, 32);
    const sSlice = byt.slice(32, 64);

    // Add the signature fields to make this a signed transaction.
    userTx.v = byt[64];
    userTx.r = ethers.BigNumber.from(rSlice).toString();
    userTx.s = ethers.BigNumber.from(sSlice).toString();
    
    // Marshal into JSON for the payload.
    var data = JSON.stringify(userTx);

    // Go doesn't want big integers to be strings. Removing quotes.
    data = data.replace('r":"', 'r":');
    data = data.replace('","s":"', ',"s":');
    data = data.replace('"}', '}');

    // Make a call to the node.
    const url = "http://localhost:8080/v1/tx/submit";
    $.post(url, data, function (o, status) {
        if ((typeof o.errors != "undefined") && (o.errors.length > 0)) {    
            window.alert("ERROR: " + o.errors[0].message);
            return;
        }

        alert("SENT");
    });
}

// =============================================================================

var formatter = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  
    // These options are needed to round to whole numbers if that's what you want.
    // minimumFractionDigits: 0, // (this suffices for whole numbers, but will print 2500.10 as $2,500.1)
    maximumFractionDigits: 0, // (causes 2500.99 to be printed as $2,501)
});

// =============================================================================

function showInfoTabSend() {
    showInfoTab("send");
}

function showInfoTabTran() {
    showInfoTab("tran");
}

function showInfoTab(which) {
    const sendBox = document.querySelector("div.sendbox");
    const tranBox = document.querySelector("div.tranbox");

    const sendBut = document.getElementById("sendbutton");
    const tranBut = document.getElementById("tranbutton");

    switch (which) {
        case "send":
            sendBox.style.display = "block";
            tranBox.style.display = "none";
            sendBut.style.backgroundColor = "#faf9f5";
            tranBut.style.backgroundColor = "#d9d8d4";
            break;
        case "tran":
            sendBox.style.display = "none";
            tranBox.style.display = "block";
            sendBut.style.backgroundColor = "#d9d8d4";
            tranBut.style.backgroundColor = "#faf9f5";
            break;
    }
}

// =============================================================================

function formatCurrencyKeyup() {
    formatCurrency($(this));
}

function formatCurrencyBlur() {
    formatCurrency($(this));
}

function formatNumber(n) {
  // format number 1000000 to 1,234,567
  return n.replace(/\D/g, "").replace(/\B(?=(\d{3})+(?!\d))/g, ",")
}

function formatCurrency(input) {
    // appends $ to value, validates decimal side
    // and puts cursor back in right position.
  
    // get input value
    var input_val = input.val();
    
    // don't validate empty input
    if (input_val === "") { return; }
    
    // original length
    var original_len = input_val.length;

    // initial caret position 
    var caret_pos = input.prop("selectionStart");

    if (input_val.indexOf(".") == 0) { return; }
    
    // no decimal entered
    // add commas to number
    // remove all non-digits
    input_val = formatNumber(input_val);
    input_val = "$" + input_val;
  
    // send updated string to input
    input.val(input_val);

    // put caret back in the right position
    var updated_len = input_val.length;
    caret_pos = updated_len - original_len + caret_pos;
    input[0].setSelectionRange(caret_pos, caret_pos);
}