const alert = document.querySelector('.alert');
alert.hidden = true;

window.onload = e => {
    // Attach click events to all copy-able texts
    const copyTargets = document.querySelectorAll('.copy');
    for (target of copyTargets) {
        target.addEventListener('click', e => {
            // Using the new Clipboard API
            if (navigator.clipboard) {
                navigator.clipboard.writeText(e.target.innerText);
            } else {
                /*
                    The Clipboard API is not fully supported yet and hence,
                    we use this junk as a fallback mechanism. (also for IE)
                    Refer: https://developer.mozilla.org/en-US/docs/Web/API/Clipboard/writeText#Browser_compatibility

                    Steps:
                    - first create an empty text area
                    - add the text from the copy target to it
                    - append the text area to the DOM tree
                    - focus and select it
                    - copy the selected text
                    - unselect everything
                    - remove the field from the DOM tree

                    Known issue:
                    On appending the text area to the DOM, the added height causes the
                    page to scroll down.
                */
                const field = document.createElement('textarea');
                field.value = e.target.innerHTML;
                document.body.appendChild(field);
                field.focus();
                field.select();
                document.execCommand('copy');
                window.getSelection().removeAllRanges();
                field.remove();
            }

            showAlert("Copied to clipboard!", 2000);
        });
    }
};

function showAlert(msg, timeout) {
    alert.innerText = msg;
    alert.hidden = false;
    
    setTimeout(() => {
        alert.hidden = true;
    }, timeout);
}