$.ajaxSetup({
    contentType: "application/json; charset=utf-8",
    error: function (xhr) {
        const conn = document.getElementById("connected");
        conn.innerHTML = "NOT CONNECTED";
      }
});

window.onload = function () {
    const url = "http://localhost:8080/v1/accounts/list/0xF01813E4B85e178A83e29B8E7bF26BD830a25f32";
    $.get(url, function (o, status) {
        const conn = document.getElementById("connected");
        conn.className = "connected";
        conn.innerHTML = "CONNECTED";

        if ((typeof o.errors != "undefined") && (o.errors.length > 0)) {    
            conn.innerHTML = "NOT CONNECTED";
            window.alert("ERROR: " + o.errors[0].message);
            return;
        }

        if (o.accounts.length != 1) {
            window.alert("ERROR: Account Not Found");
        }

        const bal = document.getElementById("balance");
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
