document.addEventListener('DOMContentLoaded', e => {
    const copyLinks = document.querySelectorAll('.copy');
    for (link of copyLinks) {
        link.addEventListener('click', e => {
            const field = document.createElement('textarea');
            field.value = e.target.innerHTML;
            document.body.appendChild(field);
            field.focus();
            field.select();
            document.execCommand('copy');
            window.getSelection().removeAllRanges();
            field.remove();
            e.target.insertAdjacentHTML('afterend', '<span class="alert alert-primary p-1 ml-2"><em>Copied!</em></span>');
            setTimeout(() => document.querySelector('.alert').remove(), 2000);
        });
    }
});