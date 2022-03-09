$.ajaxSetup({
    contentType: "application/json; charset=utf-8",
    error: function (xhr) {
        const conn = document.getElementById("connected");
        conn.innerHTML = "NOT CONNECTED";
      }
});

window.onload = function () {
    const url = "http://localhost:8080/v1/accounts/list/0xF01813E4B85e178A83e29B8E7bF26BD830a25f32";
    $.get(url, function (data, status) {
        if ((typeof data.errors != "undefined") && (data.errors.length > 0)) {
            window.alert("ERROR: " + data.errors[0].message);
            return;
        }

        const bal = document.getElementById("balance");
        bal.innerHTML = "Call Made";
    });
}
