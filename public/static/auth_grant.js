const acceptBtn = document.getElementById('acceptBtn');

acceptBtn.addEventListener('click', e => {
    const params = window.location.href.split('?')[1];
    const paramPairs = params.split('&');
    const paramsObj = {};
    
    for (const paramPair of paramPairs) {
        const pair = paramPair.split('=');
        paramsObj[pair[0]] = pair[1];
    }

    fetch('/accepted', {
        method: 'POST',
        mode: 'no-cors',
        headers: {
            'Content-Type': 'application/json'
        },
        redirect: 'follow',
        body: JSON.stringify(paramsObj)
    })
    .then(response => {
        if (response.redirected) {
            window.location.replace(response.url)
        }
    })
});