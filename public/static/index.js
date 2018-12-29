function copyText() {
    const authURLField = document.getElementById('authCodeauthURL');
    authURLField.focus();
    authURLField.select();
    document.execCommand('copy');
    window.getSelection().removeAllRanges();
    console.log('Copied!');
}