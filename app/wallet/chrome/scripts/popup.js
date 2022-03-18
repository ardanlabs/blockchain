var nonce = 0;

window.onload = function () {
    wireEvents();
    showInfoTab("send");    
    connect();
}

// =============================================================================

function wireEvents() {
    const refresh = document.getElementById("refreshsubmit");
    refresh.addEventListener(
        'click',
        load,
        false
    );

    const from = document.getElementById("from");
    from.addEventListener(
        'change',
        load,
        false
    );

    const to = document.getElementById("to");
    to.addEventListener(
        'change',
        load,
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

    const memp = document.getElementById("mempbutton");
    memp.addEventListener(
        'click',
        showInfoTabMemp,
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

    const closebutton = document.getElementById("closebutton");
    closebutton.addEventListener(
        'click',
        closeConfirmation,
        false
    );

    const confirmno = document.getElementById("confirmno");
    confirmno.addEventListener(
        'click',
        closeConfirmation,
        false
    );

    const confirmyes = document.getElementById("confirmyes");
    confirmyes.addEventListener(
        'click',
        createTransaction,
        false
    );
}

// =============================================================================

function connect() {
    var socket = new WebSocket('ws://localhost:8080/v1/events');

    socket.addEventListener('open', function (event) {
        const conn = document.getElementById("connected");
        conn.className = "connected";
        conn.innerHTML = "CONNECTED";
        load();
    });

    socket.addEventListener('close', function (event) {
        const conn = document.getElementById("connected");
        conn.className = "notconnected";
        conn.innerHTML = "NOT CONNECTED";
    });

    socket.addEventListener('message', function (event) {
        const conn = document.getElementById("connected");

        if (event.data.includes("MINING: completed")) {
            conn.className = "connected";
            conn.innerHTML = "CONNECTED";
            load();
            return;
        }

        if (event.data.includes("MINING")) {
            conn.className = "mining";
            conn.innerHTML = "MINING...";
            return;
        }
    });

    socket.addEventListener('error', function (err) {
        const conn = document.getElementById("connected");
        conn.className = "notconnected";
        conn.innerHTML = "NOT CONNECTED";
        document.getElementById("errmsg").innerText = err.message;
    });
}

// =============================================================================

$.ajaxSetup({
    contentType: "application/json; charset=utf-8",
    beforeSend: function() {
        document.getElementById("errmsg").innerText = "";
    }
});

function handleAjaxError(jqXHR, exception) {
    var msg = '';

    switch (jqXHR.status) {
    case 0:
        msg = 'Not connected, verify network.';
    case 404:
        msg = 'Requested page not found. [404]';
    case 500:
        msg = 'Internal Server Error [500].';
    default:
        switch (exception) {
        case "parsererror":
            msg = 'Requested JSON parse failed.';
        case "timeout":
            msg = 'Time out error.';
        case "abort":
            msg = 'Ajax request aborted.';
        default:
            const o = JSON.parse(jqXHR.responseText);
            msg = o.error;
        }
    }

    document.getElementById("errmsg").innerText = msg;
}

// ==============================================================================

function load() {
    const conn = document.getElementById("connected");
    if (conn.innerHTML != "CONNECTED") {
        document.getElementById("errmsg").innerText = "No connection to node.";
        return;
    }

    nonce = 0;
    document.getElementById("errmsg").innerText = "";
    document.getElementById("tranbutton").innerHTML = "Trans";

    $.ajax({
        type: "get",
        url: "http://localhost:8080/v1/genesis/list",
        success: function (response) {
            fromBalance();
            toBalance();
            transactions();
            mempool();
        },
        error: function (jqXHR, exception) {
            document.getElementById("errmsg").innerText = exception;
        },
    });
}

function fromBalance() {
    const wallet = new ethers.Wallet(document.getElementById("from").value);

    $.ajax({
        type: "get",
        url: "http://localhost:8080/v1/accounts/list/" + wallet.address,
        success: function (resp) {
            const bal = document.getElementById("frombal");
            bal.innerHTML = formatter.format(resp.accounts[0].balance) + " ARD";

            document.getElementById("fromnonce").innerHTML = resp.accounts[0].nonce;
            if (nonce == 0) {
                nonce = Number(resp.accounts[0].nonce);
                nonce += 1
            }
            document.getElementById("nextnonce").innerHTML = nonce;
        },
        error: function (jqXHR, exception) {
            handleAjaxError(jqXHR, exception);
        },
    });
}

function toBalance() {
    $.ajax({
        type: "get",
        url: "http://localhost:8080/v1/accounts/list/" + document.getElementById("to").value,
        success: function (resp) {
            const bal = document.getElementById("tobal");
            bal.innerHTML = formatter.format(resp.accounts[0].balance) + " ARD";
        },
        error: function (jqXHR, exception) {
            handleAjaxError(jqXHR, exception);
        },
    });
}

function transactions() {
    const wallet = new ethers.Wallet(document.getElementById("from").value);

    $.ajax({
        type: "get",
        url: "http://localhost:8080/v1/blocks/list/" + wallet.address,
        success: function (resp) {
            var msg = "";
            var count = 0;
            for (var i = 0; i < resp.length; i++) {
                for (var j = 0; j < resp[i].txs.length; j++) {
                    if ((resp[i].txs[j].from == wallet.address) || (resp[i].txs[j].to == wallet.address)) {
                        msg += JSON.stringify(resp[i].txs[j], null, 2);
                        count++;
                    }
                }
            }
            document.getElementById("trans").innerHTML = msg;
            document.getElementById("tranbutton").innerHTML = "Trans(" + count + ")";
        },
        error: function (jqXHR, exception) {
            handleAjaxError(jqXHR, exception);
        },
    });
}

function mempool() {
    const wallet = new ethers.Wallet(document.getElementById("from").value);

    $.ajax({
        type: "get",
        url: "http://localhost:8080/v1/tx/uncommitted/list/" + wallet.address,
        success: function (resp) {
            var msg = "";
            var count = 0;
            for (var i = 0; i < resp.length; i++) {
                msg += JSON.stringify(resp[i], null, 2);
                count++;

                if (resp[i].from == wallet.address) {

                    // Check the mempool for what the next nonce should be for this account.
                    const txNonce = Number(resp[i].nonce);
                    if (txNonce >= nonce) {
                        nonce = txNonce + 1
                        document.getElementById("nextnonce").innerHTML = nonce;
                    }

                    // Update the accounts balance.
                    const frombal = document.getElementById("frombal");
                    const txValue = Number(resp[i].value);
                    var balance = Number(frombal.innerHTML.replace(/\$|,/g, '').replace(" ARD", ""));
                    balance -= txValue;
                    frombal.innerHTML = formatter.format(balance) + " ARD";
                }
            }
            document.getElementById("mempool").innerHTML = msg;
            document.getElementById("mempbutton").innerHTML = "Mem(" + count + ")";
        },
        error: function (jqXHR, exception) {
            handleAjaxError(jqXHR, exception);
        },
    });
}

// =============================================================================

function submitTran() {
    const conn = document.getElementById("connected");
    if (conn.innerHTML != "CONNECTED") {
        document.getElementById("errmsg").innerText = "No connection to node.";
        return;
    }
    
    // Update the account information.
    fromBalance();

    // Capture and validate the amount to send.
    const amountStr = document.getElementById("sendamount").value.replace(/\$|,/g, '');
    const amount = Number(amountStr);
    if (isNaN(amount)) {
        document.getElementById("errmsg").innerText = "Amount is not a number.";
        return;
    }
    if (amount <= 0) {
        document.getElementById("errmsg").innerText = "Amount must be greater than 0 dollars.";
        return;
    }

    // Capture and validate the tip to send.
    const tipStr = document.getElementById("sendtip").value.replace(/\$|,/g, '');
    const tip = Number(tipStr);
    if (isNaN(tip)) {
        document.getElementById("errmsg").innerText = "Tip is not a number.";
        return;
    }
    if (tip < 0) {
        document.getElementById("errmsg").innerText = "Tip can't be a negative number.";
        return;
    }

    // Validate there is enough money.
    const frombal = document.getElementById("frombal");
    var balance = Number(frombal.innerHTML.replace(/\$|,/g, '').replace(" ARD", ""));
    if (amount > balance) {
        document.getElementById("errmsg").innerText = "You don't have enough money.";
        return;
    }

    showConfirmation();
}

function createTransaction() {

    // We got a yes confirmation so we know the values are verified.
    const amountStr = document.getElementById("sendamount").value.replace(/\$|,/g, '');
    const tipStr = document.getElementById("sendtip").value.replace(/\$|,/g, '');

     // Construct a userTx with all the information.
     const userTx = {
        nonce: nonce,
        to: document.getElementById("to").value,
        value: Number(amountStr),
        tip: Number(tipStr),
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

    closeConfirmation();

    $.ajax({
        type: "post",
        url: "http://localhost:8080/v1/tx/submit",
        data: data,
        success: function (resp) {
            document.getElementById("nextnonce").innerHTML = nonce;
            load();
        },
        error: function (jqXHR, exception) {
            handleAjaxError(jqXHR, exception);
        },
    });
}

// =============================================================================

function showConfirmation() {
    const modal = document.getElementById("confirmationmodal");
    modal.style.display = "block";

    document.getElementById("yesnomessage").innerHTML = "";
}

function closeConfirmation() {
    const modal = document.getElementById("confirmationmodal");
    modal.style.display = "none";
}

function onConfirm() {

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

function showInfoTabMemp() {
    showInfoTab("memp");
}

function showInfoTab(which) {
    const sendBox = document.querySelector("div.sendbox");
    const tranBox = document.querySelector("div.tranbox");
    const mempBox = document.querySelector("div.mempbox");

    const sendBut = document.getElementById("sendbutton");
    const tranBut = document.getElementById("tranbutton");
    const mempBut = document.getElementById("mempbutton");

    switch (which) {
    case "send":
        sendBox.style.display = "block";
        tranBox.style.display = "none";
        mempBox.style.display = "none";
        sendBut.style.backgroundColor = "#faf9f5";
        tranBut.style.backgroundColor = "#d9d8d4";
        mempBut.style.backgroundColor = "#d9d8d4";
        break;
    case "tran":
        tranBox.style.display = "block";
        sendBox.style.display = "none";
        mempBox.style.display = "none";
        tranBut.style.backgroundColor = "#faf9f5";
        sendBut.style.backgroundColor = "#d9d8d4";
        mempBut.style.backgroundColor = "#d9d8d4";
        break;
    case "memp":
        mempBox.style.display = "block";
        tranBox.style.display = "none";
        sendBox.style.display = "none";
        mempBut.style.backgroundColor = "#faf9f5";
        tranBut.style.backgroundColor = "#d9d8d4";
        sendBut.style.backgroundColor = "#d9d8d4";
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