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

// Check if the redirect URI is passed in the URL
var redirectURI = parseParams()["redirect_uri"];
const node = document.getElementById("redirectURI");

// If redirect URI not passed and if using Implicit flow,
// unhide the redirectURI field.
// Else, just set the value, but don't unhide.
if (redirectURI === undefined) {
    redirectURI = "";
    node.hidden = false;
} else {
    node.value = redirectURI;
}
