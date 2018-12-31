const acceptBtn = document.getElementById('acceptBtn');
const cancelBtn = document.getElementById('cancelBtn');

acceptBtn.addEventListener('click', e => {
    fetch('/accepted', {
        method: 'POST',
        mode: 'no-cors',
        headers: {
            'Content-Type': 'application/json'
        },
        redirect: 'follow',
        body: JSON.stringify(parseParams())
    })
    .then(response => {
        if (response.redirected) {
            window.location.replace(response.url)
        }
    })
});

cancelBtn.addEventListener('click', e => {
    redirectURI = parseParams().redirect_uri + "?error=access_denied" 
    window.location.replace(redirectURI)
});

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