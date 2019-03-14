function parseParams(url) {
    const params = window.location.href.split('?')[1];
    const paramPairs = params.split('&');
    const paramsObj = {};
    
    for (const paramPair of paramPairs) {
        const pair = paramPair.split('=');
        paramsObj[pair[0]] = pair[1];
    }

    return paramsObj;
}

const redirectURI = parseParams()["redirect_uri"];
const node = document.getElementById("redirectURI");
node.value = redirectURI;