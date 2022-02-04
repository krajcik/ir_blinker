window.addEventListener("load", function () {
    let connInteval;
    console.log(ws_addr);
    let ws = new WebSocket(ws_addr);
    ws.onopen = function () {
        console.log("Websocket connection opened");
        clearInterval(connInteval);
        ws.send("");
    }
    ws.onclose = function () {
        console.error("Websocket connection closed");
        connInteval = setTimeout(function () {
            _ws = ws
            ws = new WebSocket(ws_addr);
            ws.onopen = _ws.onopen
            ws.onclose = _ws.onclose
            ws.onerror = _ws.onerror
            ws.onmessage = _ws.onmessage
            _ws = null
            console.log("Connection lost, reconnecting...");
        }, 5000)
        showFlash("Connection lost, reconnecting...");
    }
    ws.onerror = function (evt) {
        console.error("Websocket error.", evt.data);
    }
    ws.onmessage = function (evt) {
        let data = JSON.parse(evt.data);
        if (!data.IsConnected) {
            showFlash("Waiting for iRacing connection...")
            return;
        }
        document.getElementById("flash").style.display = "none";

        rpmLight(data.RPMLights, data.RPM.toFixed(0), data.Gear)
        pitLimiter(data.EngineWarnings)
        absLight(data.AbsActive)
    }
});

function showFlash(msg) {
    document.getElementById("flash").innerHTML = msg;
    document.getElementById("flash").style.display = "block";
}

function rpmLight(limits, rpm, gear) {
    document.getElementById("overlay").classList.remove("change", "blink")
    if (limits.HasGears) {
        up = limits.Gears[gear-1];
        if (parseFloat(rpm) >= parseFloat(up)){
            document.getElementById("overlay").classList.add("blink");
            return;
        }
        if (parseFloat(rpm) >= (parseFloat(up) * 0.90)){
            document.getElementById("overlay").classList.add("change");
            return;
        }
    }
    if (parseFloat(rpm) >= parseFloat(limits.Blink)) {
        document.getElementById("overlay").classList.add("blink");
        return;
    }
    if (parseFloat(rpm) >= parseFloat(limits.Last)) {
        document.getElementById("overlay").classList.add("change");
        return;
    }

    if (parseFloat(rpm) >= parseFloat(limits.Shift)) {
        document.getElementById("overlay").classList.add("shift");
    }
}

function absLight(isActive) {
    document.getElementById("overlay").classList.remove("abs")
    if (isActive) {
        document.getElementById("overlay").classList.add("abs")
    }
}

function pitLimiter(engineWarnings) {
    if (0 === engineWarnings.indexOf("0x1") || 0 === engineWarnings.indexOf("0x3")) {
        document.getElementById("overlay").className = "pit-limiter";
    } else {
        document.getElementById("overlay").classList.remove("pit-limiter");
    }
}