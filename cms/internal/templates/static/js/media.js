(function () {
  'use strict';

  const form = document.getElementById('media-upload-form');
  const status = document.getElementById('upload-status');

  if (form) {
    form.addEventListener('submit', async function (e) {
      e.preventDefault();
      const fileInput = document.getElementById('media-file');
      if (!fileInput.files[0]) return;

      const formData = new FormData();
      formData.append('file', fileInput.files[0]);

      status.textContent = 'Uploading…';
      status.className = 'save-status saving';

      try {
        const res = await fetch('/api/media/upload', {
          method: 'POST',
          body: formData,
        });
        const result = await res.json();
        if (!res.ok) throw new Error(result.error || 'Upload failed');

        status.textContent = 'Uploaded!';
        status.className = 'save-status saved';
        fileInput.value = '';
        setTimeout(function () { window.location.reload(); }, 800);
      } catch (err) {
        status.textContent = err.message;
        status.className = 'save-status error';
      }
    });
  }

  document.querySelectorAll('.copy-md').forEach(function (btn) {
    btn.addEventListener('click', function () {
      const url = btn.getAttribute('data-url');
      const md = '![](' + url + ')';
      navigator.clipboard.writeText(md).then(function () {
        btn.textContent = 'Copied!';
        setTimeout(function () { btn.textContent = 'Copy Markdown'; }, 1500);
      });
    });
  });
})();
