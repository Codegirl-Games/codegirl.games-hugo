(function () {
  'use strict';

  const AUTOSAVE_INTERVAL = 30000;
  const DRAFT_KEY_PREFIX = 'cms-draft-';

  let editor;
  let dirty = false;
  let autosaveTimer;

  function slugify(text) {
    return text
      .toLowerCase()
      .trim()
      .replace(/[^\w\s-]/g, '')
      .replace(/[\s_]+/g, '-')
      .replace(/-+/g, '-')
      .replace(/^-|-$/g, '');
  }

  function getFormData() {
    return {
      title: document.getElementById('title').value,
      slug: document.getElementById('slug').value,
      date: document.getElementById('date').value,
      draft: document.getElementById('draft').checked,
      tags: document.getElementById('tags').value,
      body: editor ? editor.value() : document.getElementById('body').value,
      original: document.getElementById('original').value,
    };
  }

  function setStatus(text, className) {
    const el = document.getElementById('save-status');
    if (!el) return;
    el.textContent = text;
    el.className = 'save-status ' + (className || '');
  }

  function draftKey() {
    const original = document.getElementById('original').value;
    const slug = document.getElementById('slug').value;
    return DRAFT_KEY_PREFIX + (original || slug || 'new');
  }

  function saveDraftLocal() {
    try {
      localStorage.setItem(draftKey(), JSON.stringify(getFormData()));
    } catch (_) { /* quota exceeded */ }
  }

  function loadDraftLocal() {
    try {
      const raw = localStorage.getItem(draftKey());
      if (!raw) return;
      const data = JSON.parse(raw);
      if (!confirm('A local autosave draft was found. Restore it?')) {
        localStorage.removeItem(draftKey());
        return;
      }
      document.getElementById('title').value = data.title || '';
      document.getElementById('slug').value = data.slug || '';
      document.getElementById('date').value = data.date || '';
      document.getElementById('draft').checked = !!data.draft;
      document.getElementById('tags').value = data.tags || '';
      if (editor) editor.value(data.body || '');
      dirty = true;
    } catch (_) { /* ignore */ }
  }

  function clearDraftLocal() {
    localStorage.removeItem(draftKey());
  }

  async function savePost() {
    const data = getFormData();
    setStatus('Saving…', 'saving');

    try {
      const res = await fetch('/api/posts/save', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });
      const result = await res.json();
      if (!res.ok) throw new Error(result.error || 'Save failed');

      dirty = false;
      clearDraftLocal();
      setStatus('Saved', 'saved');

      if (result.slug && result.slug !== data.original) {
        document.getElementById('original').value = result.slug;
        history.replaceState(null, '', '/admin/posts/' + result.slug);
      }

      setTimeout(function () {
        if (!dirty) setStatus('', '');
      }, 2000);
    } catch (err) {
      setStatus(err.message, 'error');
    }
  }

  async function uploadImage(file) {
    const form = new FormData();
    form.append('file', file);

    const res = await fetch('/api/media/upload', {
      method: 'POST',
      body: form,
    });
    const result = await res.json();
    if (!res.ok) throw new Error(result.error || 'Upload failed');
    return result;
  }

  function insertMarkdown(text) {
    if (!editor) return;
    const cm = editor.codemirror;
    const doc = cm.getDoc();
    const cursor = doc.getCursor();
    doc.replaceRange(text, cursor);
    dirty = true;
  }

  async function handleImageUpload(file) {
    try {
      setStatus('Uploading image…', 'saving');
      const result = await uploadImage(file);
      insertMarkdown('![](' + result.url + ')');
      setStatus('Image uploaded', 'saved');
      setTimeout(function () { if (!dirty) setStatus('', ''); }, 2000);
    } catch (err) {
      setStatus(err.message, 'error');
    }
  }

  function openMediaModal() {
    const modal = document.getElementById('media-modal');
    const grid = document.getElementById('media-grid');
    modal.classList.remove('hidden');
    grid.innerHTML = '<p>Loading…</p>';

    fetch('/api/media')
      .then(function (r) { return r.json(); })
      .then(function (data) {
        grid.innerHTML = '';
        if (!data.items || !data.items.length) {
          grid.innerHTML = '<p class="empty-state">No images yet.</p>';
          return;
        }
        data.items.forEach(function (item) {
          const div = document.createElement('div');
          div.className = 'media-item';
          div.innerHTML = '<img src="' + item.url + '" alt="' + item.name + '">';
          div.addEventListener('click', function () {
            insertMarkdown('![](' + item.url + ')');
            modal.classList.add('hidden');
          });
          grid.appendChild(div);
        });
      })
      .catch(function () {
        grid.innerHTML = '<p class="empty-state">Failed to load media.</p>';
      });
  }

  function initEditor() {
    const textarea = document.getElementById('body');
    if (!textarea || typeof EasyMDE === 'undefined') return;

    editor = new EasyMDE({
      element: textarea,
      autofocus: true,
      spellChecker: false,
      autosave: { enabled: false },
      toolbar: [
        'bold', 'italic', 'heading', '|',
        'quote', 'unordered-list', 'ordered-list', '|',
        'link', 'image', '|',
        'preview', 'side-by-side', 'fullscreen', '|',
        'guide',
      ],
      status: ['lines', 'words'],
      renderingConfig: { singleLineBreaks: false },
      uploadImage: true,
      imageUploadFunction: function (file, onSuccess, onError) {
        uploadImage(file)
          .then(function (r) { onSuccess(r.url); })
          .catch(function (e) { onError(e.message); });
      },
    });

    editor.codemirror.on('change', function () {
      dirty = true;
    });

    // Clipboard image paste
    editor.codemirror.getWrapperElement().addEventListener('paste', function (e) {
      const items = e.clipboardData && e.clipboardData.items;
      if (!items) return;
      for (let i = 0; i < items.length; i++) {
        if (items[i].type.indexOf('image') !== -1) {
          e.preventDefault();
          handleImageUpload(items[i].getAsFile());
          return;
        }
      }
    });
  }

  function initSlugGeneration() {
    const title = document.getElementById('title');
    const slug = document.getElementById('slug');
    const original = document.getElementById('original').value;
    let slugManual = !!original;

    slug.addEventListener('input', function () {
      slugManual = true;
    });

    title.addEventListener('input', function () {
      if (!slugManual) {
        slug.value = slugify(title.value);
      }
      dirty = true;
    });

    ['slug', 'date', 'tags', 'draft'].forEach(function (id) {
      const el = document.getElementById(id);
      if (el) el.addEventListener('change', function () { dirty = true; });
      if (el) el.addEventListener('input', function () { dirty = true; });
    });
  }

  function initButtons() {
    document.getElementById('save-btn').addEventListener('click', savePost);

    document.getElementById('insert-media-btn').addEventListener('click', openMediaModal);

    document.getElementById('upload-image-btn').addEventListener('click', function () {
      document.getElementById('image-upload').click();
    });

    document.getElementById('image-upload').addEventListener('change', function (e) {
      if (e.target.files[0]) handleImageUpload(e.target.files[0]);
      e.target.value = '';
    });

    document.querySelectorAll('[data-close-modal]').forEach(function (el) {
      el.addEventListener('click', function () {
        document.getElementById('media-modal').classList.add('hidden');
      });
    });

    // Keyboard shortcut: Ctrl/Cmd+S
    document.addEventListener('keydown', function (e) {
      if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        e.preventDefault();
        savePost();
      }
    });
  }

  function initAutosave() {
    autosaveTimer = setInterval(function () {
      if (dirty) saveDraftLocal();
    }, AUTOSAVE_INTERVAL);
  }

  function initUnsavedWarning() {
    window.addEventListener('beforeunload', function (e) {
      if (dirty) {
        e.preventDefault();
        e.returnValue = '';
      }
    });
  }

  document.addEventListener('DOMContentLoaded', function () {
    initEditor();
    initSlugGeneration();
    initButtons();
    initAutosave();
    initUnsavedWarning();
    loadDraftLocal();
  });
})();
