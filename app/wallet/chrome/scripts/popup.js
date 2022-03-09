$.ajaxSetup({
    contentType: "application/json; charset=utf-8",
    error: function (xhr) {
        const conn = document.getElementById("connected");
        conn.innerHTML = "NOT CONNECTED";
      }
});

window.onload = function () {
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

    connect();
}

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

function fromBalance() {
    const url = "http://localhost:8080/v1/accounts/list/" + document.getElementById("from").value;

    $.get(url, function (o, status) {
        if ((typeof o.errors != "undefined") && (o.errors.length > 0)) {    
            window.alert("ERROR: " + o.errors[0].message);
            return;
        }

        const bal = document.getElementById("frombal");
        bal.innerHTML = formatter.format(o.accounts[0].balance);
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
        bal.innerHTML = formatter.format(o.accounts[0].balance);
    });
}

var formatter = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  
    // These options are needed to round to whole numbers if that's what you want.
    //minimumFractionDigits: 0, // (this suffices for whole numbers, but will print 2500.10 as $2,500.1)
    //maximumFractionDigits: 0, // (causes 2500.99 to be printed as $2,501)
  });
