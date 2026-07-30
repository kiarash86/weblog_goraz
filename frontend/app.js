/* ==========================================================================
   Marginalia — app logic
   Vanilla JS, no build step. See config.js for the real API contract this
   backend exposes (auth is bearer-token only, boards have no created_at,
   etc). Both the feed and the single-board detail endpoint include
   author_username now (see boardAuthorLabel).
   ========================================================================== */

const API_BASE = (window.MARGINALIA_CONFIG && window.MARGINALIA_CONFIG.API_BASE) || '';
const STORAGE_KEY = 'marginalia_auth';

function loadStoredAuth() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch (_) {
    return null;
  }
}

function saveStoredAuth(token, user) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ token, user }));
  } catch (_) { /* ignore */ }
}

function clearStoredAuth() {
  try { localStorage.removeItem(STORAGE_KEY); } catch (_) { /* ignore */ }
}

const stored = loadStoredAuth();
let AUTH_TOKEN = stored ? stored.token : null;

const FEED_PAGE_SIZE = 10; // matches the backend's fixed LIMIT per page

const state = {
  user: stored ? stored.user : null, // { id, username } | null
  route: { name: 'feed' }, // {name:'feed'} | {name:'detail', id}
  boards: [],
  detail: null,
  comments: [],
  loading: true,
  error: null,
  page: 1,
  search: '', // raw text the user typed, no wildcards
  hasNextPage: false, // inferred: true if the last page came back full
};

const root = document.getElementById('main-root');
const headerSlot = document.getElementById('header-user-slot');
const modalRoot = document.getElementById('modal-root');

/* ---------------------------- API wrapper ---------------------------- */

async function api(path, { method = 'GET', body, returnResponse = false } = {}) {
  const headers = { 'Content-Type': 'application/json' };
  if (AUTH_TOKEN) headers['Authorization'] = `Bearer ${AUTH_TOKEN}`;

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (res.status === 204) return null;

  let data = null;
  try { data = await res.json(); } catch (_) { /* empty body */ }

  if (res.status === 401) {
    // Token missing/expired/invalid — the backend has no /me to pre-check
    // this, so we only find out when a real request fails.
    AUTH_TOKEN = null;
    state.user = null;
    clearStoredAuth();
  }

  if (!res.ok) {
    const msg = (data && (data.message || data.error)) || `Request failed (${res.status})`;
    throw new Error(msg);
  }
  // Some endpoints (feed pagination) need a response header, not just the
  // parsed body — opt in per call so every other call site is unaffected.
  return returnResponse ? { data, res } : data;
}

async function uploadImage(file) {
  const headers = {};
  if (AUTH_TOKEN) headers['Authorization'] = `Bearer ${AUTH_TOKEN}`;
  // NOTE: no Content-Type header here on purpose — the browser sets the
  // multipart boundary itself. Setting it manually breaks the upload.

  const formData = new FormData();
  formData.append('image', file);

  const res = await fetch(`${API_BASE}/upload`, {
    method: 'POST',
    headers,
    body: formData,
  });

  let data = null;
  try { data = await res.json(); } catch (_) { /* empty body */ }

  if (!res.ok) {
    const msg = (data && (data.message || data.error)) || `Upload failed (${res.status})`;
    throw new Error(msg);
  }
  return data.path;
}

/* ---------------------------- helpers ---------------------------- */

function escapeHtml(str) {
  return String(str ?? '').replace(/[&<>"']/g, (c) => ({
    '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;',
  }[c]));
}

function el(html) {
  const t = document.createElement('template');
  t.innerHTML = html.trim();
  return t.content.firstElementChild;
}

// Fallback for when we only have an author_id and no username to go with
// it. We only know our own logged-in username locally, so that's the one
// case we can label with a handle; everyone else shows as "User #<id>".
function authorLabel(authorId) {
  if (state.user && state.user.id === authorId) return `@${escapeHtml(state.user.username)}`;
  return `User #${authorId}`;
}

// Boards from the feed (GET /weblog) come back with an author_username
// field from the backend, so we show the real handle instead of the
// numeric id.
function boardAuthorLabel(board) {
  if (board.author_username) return `@${escapeHtml(board.author_username)}`;
  return authorLabel(board.author_id);
}

// Single-letter avatar used on feed cards. Falls back to "?" when we have
// neither a username nor (for someone else's board) any way to resolve one.
function boardAuthorInitial(board) {
  const uname = board.author_username
    || (state.user && state.user.id === board.author_id ? state.user.username : null);
  return uname ? escapeHtml(uname.charAt(0).toUpperCase()) : '?';
}

// Comments come back with an author_username field from the backend now,
// so we can show the real handle instead of "User #<id>".
function commentAuthorLabel(comment) {
  if (comment.author_username) return `@${escapeHtml(comment.author_username)}`;
  return authorLabel(comment.author_id);
}

/* ---------------------------- render: header ---------------------------- */

function renderHeader() {
  headerSlot.innerHTML = '';
  if (state.user) {
    headerSlot.appendChild(el(`
      <div style="display:flex; align-items:center; gap:12px;">
        <span class="handle">@${escapeHtml(state.user.username)}</span>
        <button class="btn btn-brass" id="btn-new-entry">+ New entry</button>
        <button class="btn btn-ghost" id="btn-logout">Log out</button>
      </div>
    `));
    headerSlot.querySelector('#btn-new-entry').onclick = openNewEntryModal;
    headerSlot.querySelector('#btn-logout').onclick = logout;
  } else {
    headerSlot.appendChild(el(`<span style="color:var(--paper-dim); font-size:12px;">the notebook everyone's writing in</span>`));
  }
}

/* ---------------------------- render: auth ---------------------------- */

function renderAuth() {
  root.innerHTML = '';
  const wrap = el(`
    <div class="auth-shell">
      <div class="auth-card">
        <div class="tabs">
          <button class="tab active" data-tab="login">Log in</button>
          <button class="tab" data-tab="signup">Sign up</button>
        </div>
        <h1 class="auth-title" id="auth-title">Welcome back</h1>
        <p class="auth-sub" id="auth-sub">Log in to read, write, and leave notes in the margin.</p>
        <div id="auth-error"></div>
        <form id="auth-form">
          <div class="field">
            <label for="f-username">Username</label>
            <input type="text" id="f-username" autocomplete="username" required>
          </div>
          <div class="field">
            <label for="f-password">Password</label>
            <input type="password" id="f-password" autocomplete="current-password" required>
          </div>
          <button type="submit" class="btn btn-brass btn-block" id="auth-submit">Log in</button>
        </form>
      </div>
    </div>
  `);
  root.appendChild(wrap);

  let mode = 'login';
  const tabs = wrap.querySelectorAll('.tab');
  const title = wrap.querySelector('#auth-title');
  const sub = wrap.querySelector('#auth-sub');
  const submitBtn = wrap.querySelector('#auth-submit');
  const errSlot = wrap.querySelector('#auth-error');

  function setMode(m) {
    mode = m;
    tabs.forEach(t => t.classList.toggle('active', t.dataset.tab === m));
    if (m === 'login') {
      title.textContent = 'Welcome back';
      sub.textContent = 'Log in to read, write, and leave notes in the margin.';
      submitBtn.textContent = 'Log in';
    } else {
      title.textContent = 'Start a notebook';
      sub.textContent = 'Pick a username — this is how others will see your entries and notes.';
      submitBtn.textContent = 'Sign up';
    }
    errSlot.innerHTML = '';
  }
  tabs.forEach(t => t.onclick = () => setMode(t.dataset.tab));

  wrap.querySelector('#auth-form').onsubmit = async (e) => {
    e.preventDefault();
    const username = wrap.querySelector('#f-username').value.trim();
    const password = wrap.querySelector('#f-password').value;
    submitBtn.disabled = true;
    errSlot.innerHTML = '';
    try {
      // Both /signup and /login return { user: <id>, token } directly —
      // signup already logs you in, no separate login call needed.
      const res = mode === 'signup'
        ? await api('/signup', { method: 'POST', body: { username, password } })
        : await api('/login', { method: 'POST', body: { username, password } });

      AUTH_TOKEN = res.token;
      state.user = { id: res.user, username };
      saveStoredAuth(AUTH_TOKEN, state.user);
      await loadFeed();
      renderHeader();
      renderRoute();
    } catch (err) {
      errSlot.appendChild(el(`<div class="banner banner-error">${escapeHtml(err.message)}</div>`));
    } finally {
      submitBtn.disabled = false;
    }
  };
}

/* ---------------------------- render: feed ---------------------------- */

function stampMarkup(isPrivate) {
  const cls = isPrivate ? 'private' : 'public';
  const label = isPrivate ? 'Private' : 'Public';
  return `
    <div class="stamp ${cls}">
      <b>${label}</b>
    </div>
  `;
}

function renderFeed() {
  // Was the search box focused before this re-render? Re-renders happen
  // on every keystroke (via loadFeed), pagination click, and initial
  // load — we only want to restore focus/caret for the first case.
  const wasSearchFocused = document.activeElement && document.activeElement.id === 'feed-search';
  const caretPos = wasSearchFocused ? document.activeElement.selectionStart : null;

  root.innerHTML = '';
  const wrap = el(`<div></div>`);

  wrap.appendChild(el(`
    <div class="feed-toolbar">
      <div class="feed-heading">
        <span class="feed-eyebrow">Marginalia</span>
        <h1 class="feed-title">The Feed</h1>
      </div>
      <div class="search-wrap">
        <svg viewBox="0 0 20 20" fill="none" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="8.5" cy="8.5" r="6"></circle>
          <line x1="17" y1="17" x2="13.2" y2="13.2"></line>
        </svg>
        <input type="search" class="search-input" id="feed-search" placeholder="Search titles…" value="${escapeHtml(state.search)}">
      </div>
    </div>
  `));
  wrap.appendChild(el(`<div class="feed-rule"></div>`));

  const searchInput = wrap.querySelector('#feed-search');
  let searchDebounce;
  searchInput.oninput = () => {
    clearTimeout(searchDebounce);
    const value = searchInput.value;
    searchDebounce = setTimeout(() => {
      state.search = value.trim();
      state.page = 1;
      loadFeed();
    }, 700);
  };
  if (wasSearchFocused) {
    searchInput.focus();
    searchInput.setSelectionRange(caretPos, caretPos);
  }

  if (state.error) {
    wrap.appendChild(el(`<div class="banner banner-error">${escapeHtml(state.error)}</div>`));
  }

  if (state.loading) {
    wrap.appendChild(el(`<div class="loading-row">Turning the pages…</div>`));
    root.appendChild(wrap);
    return;
  }

  if (!state.boards.length) {
    wrap.appendChild(el(`
      <div class="empty-state">
        <div class="stamp public" style="margin:0 auto 16px;"><b>Empty</b></div>
        <h3>${state.page > 1 ? 'Nothing on this page' : (state.search ? 'No entries match your search' : 'Nothing logged yet')}</h3>
        <p>${state.page > 1 ? 'You\'ve gone past the last entry.' : (state.search ? 'Try a different search term.' : 'Be the first to write an entry.')}</p>
      </div>
    `));
    if (state.page > 1) {
      wrap.appendChild(buildPagination());
    }
    root.appendChild(wrap);
    return;
  }

  const list = el(`<div class="entry-list"></div>`);
  state.boards.forEach((b, i) => {
    const indexLabel = String((state.page - 1) * FEED_PAGE_SIZE + i + 1).padStart(3, '0');
    const card = el(`
      <article class="entry-card transition">
        <div class="entry-index">${indexLabel}</div>
        <div class="entry-content">
          <div class="entry-meta-row">
            <span class="entry-meta">
              <span class="author-avatar">${boardAuthorInitial(b)}</span>
              <span class="author">${boardAuthorLabel(b)}</span>
            </span>
            ${stampMarkup(b.is_private)}
          </div>
          <h3>${escapeHtml(b.title)}</h3>
          <p class="entry-excerpt">${escapeHtml((b.content || '').slice(0, 160))}${(b.content || '').length > 160 ? '…' : ''}</p>
        </div>
      </article>
    `);
    card.onclick = () => navigate({ name: 'detail', id: b.id });
    list.appendChild(card);
  });
  wrap.appendChild(list);
  wrap.appendChild(buildPagination());

  root.appendChild(wrap);
}

function buildPagination() {
  const pagination = el(`
    <div class="pagination">
      <button class="btn btn-ghost" id="btn-prev-page">← Prev</button>
      <span class="pagination-page">Page ${state.page}</span>
      <button class="btn btn-ghost" id="btn-next-page">Next →</button>
    </div>
  `);
  const prevBtn = pagination.querySelector('#btn-prev-page');
  const nextBtn = pagination.querySelector('#btn-next-page');
  prevBtn.disabled = state.page <= 1;
  nextBtn.disabled = !state.hasNextPage;
  prevBtn.onclick = () => {
    if (state.page <= 1) return;
    state.page -= 1;
    loadFeed();
  };
  nextBtn.onclick = () => {
    if (!state.hasNextPage) return;
    state.page += 1;
    loadFeed();
  };
  return pagination;
}

async function loadFeed() {
  state.loading = true;
  state.error = null;
  renderFeed();
  try {
    const params = new URLSearchParams({ page: String(state.page) });
    // The backend does a raw ILIKE with whatever we send, so wrap the
    // user's term in wildcards for partial matching; an empty term is
    // left out and the backend defaults to matching everything.
    if (state.search) params.set('search', `%${state.search}%`);
    const { data: boards, res } = await api(`/weblog?${params.toString()}`, { returnResponse: true });
    state.boards = boards || [];
    // The backend now tells us directly (via X-Has-Next) whether a real
    // next page exists, instead of us guessing from a full page — a full
    // page doesn't always mean there's more (e.g. exactly 10 boards total).
    state.hasNextPage = res.headers.get('Next-Page') === 'true';
  } catch (err) {
    state.error = err.message;
    state.boards = [];
    state.hasNextPage = false;
    if (!state.user) { renderAuth(); return; }
  } finally {
    state.loading = false;
    renderFeed();
  }
}

/* ---------------------------- render: detail ---------------------------- */

async function renderDetail(id) {
  root.innerHTML = '';
  root.appendChild(el(`<div class="loading-row">Opening the page…</div>`));
  let b, comments;
  try {
    // The backend has no "board with comments embedded" response — these
    // are two separate calls (GET /weblog/:id and GET /weblog/:id/comment).
    [b, comments] = await Promise.all([
      api(`/weblog/${id}`),
      api(`/weblog/${id}/comment`).catch(() => []), // tolerate empty/absent list
    ]);
    state.detail = b;
    state.comments = comments || [];
  } catch (err) {
    root.innerHTML = '';
    root.appendChild(el(`
      <a href="#" class="back-link" id="back">← Back to feed</a>
      <div class="banner banner-error">${escapeHtml(err.message)}</div>
    `));
    root.querySelector('#back').onclick = (e) => { e.preventDefault(); navigate({ name: 'feed' }); };
    return;
  }

  const isOwner = state.user && state.user.id === b.author_id;

  root.innerHTML = '';
  const wrap = el(`
    <div class="entry-detail">
      <a href="#" class="back-link" id="back">← Back to feed</a>
      <div class="entry-header">
        <h1>${escapeHtml(b.title)}</h1>
        ${stampMarkup(b.is_private)}
      </div>
      <span class="entry-meta">${boardAuthorLabel(b)}</span>
      ${b.img_path ? `<img class="entry-image" src="${escapeHtml(b.img_path)}" alt="">` : ''}
      <p class="entry-body">${escapeHtml(b.content)}</p>
      <div class="entry-actions">
        ${isOwner ? `<button class="btn btn-brass" id="btn-share">Share</button>` : ''}
        ${isOwner ? `<button class="btn btn-clay" id="btn-delete">Delete entry</button>` : ''}
      </div>
      <div class="margin-header">
        <h2>Marginalia</h2>
        <span class="margin-count">${state.comments.length} note${state.comments.length === 1 ? '' : 's'}</span>
      </div>
      <div id="notes-list"></div>
      <div id="notes-form-slot"></div>
    </div>
  `);
  root.appendChild(wrap);

  wrap.querySelector('#back').onclick = (e) => { e.preventDefault(); navigate({ name: 'feed' }); };
  if (isOwner) {
    wrap.querySelector('#btn-delete').onclick = () => deleteBoard(b.id);
    wrap.querySelector('#btn-share').onclick = () => openShareModal(b.id);
  }

  const notesList = wrap.querySelector('#notes-list');
  state.comments.forEach(c => {
    const canDelete = state.user && state.user.id === c.author_id;
    const note = el(`
      <div class="note">
        <div class="note-meta">${commentAuthorLabel(c)}${canDelete ? ' <button class="btn-ghost" style="border:none;padding:0 0 0 8px;text-transform:none;font-size:11px;" data-del>remove</button>' : ''}</div>
        <p class="note-text">${escapeHtml(c.content)}</p>
      </div>
    `);
    if (canDelete) {
      note.querySelector('[data-del]').onclick = () => deleteComment(b.id, c.id);
    }
    notesList.appendChild(note);
  });

  const formSlot = wrap.querySelector('#notes-form-slot');
  if (state.user) {
    const form = el(`
      <form class="note-form" id="note-form">
        <textarea id="note-text" placeholder="Write a note in the margin…" required></textarea>
        <div class="note-form-footer">
          <button type="submit" class="btn btn-brass">Add note</button>
        </div>
      </form>
    `);
    form.onsubmit = async (e) => {
      e.preventDefault();
      const textarea = form.querySelector('#note-text');
      const content = textarea.value.trim();
      if (!content) return;
      const btn = form.querySelector('button');
      btn.disabled = true;
      try {
        await api(`/weblog/${b.id}/comment`, { method: 'POST', body: { content } });
        textarea.value = '';
        renderDetail(b.id);
      } catch (err) {
        form.insertAdjacentElement('beforebegin', el(`<div class="banner banner-error">${escapeHtml(err.message)}</div>`));
      } finally {
        btn.disabled = false;
      }
    };
    formSlot.appendChild(form);
  } else {
    formSlot.appendChild(el(`<div class="signin-prompt">Log in to leave a note in the margin.</div>`));
  }
}

async function deleteComment(boardId, commentId) {
  if (!confirm('Remove this note?')) return;
  try {
    await api(`/weblog/${boardId}/comment/${commentId}`, { method: 'DELETE' });
    renderDetail(boardId);
  } catch (err) {
    alert(err.message);
  }
}

async function deleteBoard(id) {
  if (!confirm('Delete this entry? This can\'t be undone.')) return;
  try {
    await api(`/weblog/${id}`, { method: 'DELETE' });
    await loadFeed();
    navigate({ name: 'feed' });
  } catch (err) {
    alert(err.message);
  }
}

/* ---------------------------- share modal ---------------------------- */

function openShareModal(boardId) {
  modalRoot.innerHTML = '';
  const overlay = el(`
    <div class="modal-overlay">
      <div class="modal-card">
        <h2 class="modal-title">Share this entry</h2>
        <div id="share-error"></div>
        <form id="share-form">
          <div class="field">
            <label for="s-usernames">Share with (usernames)</label>
            <input type="text" id="s-usernames" placeholder="alice, bob" required>
            <span class="field-hint">Comma-separated usernames.</span>
          </div>
          <div class="modal-actions">
            <button type="button" class="btn btn-ghost" id="btn-cancel">Cancel</button>
            <button type="submit" class="btn btn-brass" id="btn-share-submit">Share</button>
          </div>
        </form>
      </div>
    </div>
  `);
  modalRoot.appendChild(overlay);
  overlay.querySelector('#btn-cancel').onclick = () => modalRoot.innerHTML = '';
  overlay.onclick = (e) => { if (e.target === overlay) modalRoot.innerHTML = ''; };

  overlay.querySelector('#share-form').onsubmit = async (e) => {
    e.preventDefault();
    const errSlot = overlay.querySelector('#share-error');
    const btn = overlay.querySelector('#btn-share-submit');
    const usernames = overlay.querySelector('#s-usernames').value
      .split(',').map(s => s.trim()).filter(Boolean);
    if (!usernames.length) return;
    btn.disabled = true;
    errSlot.innerHTML = '';
    try {
      // Response keys come straight from the backend as written:
      // { "founded seccusfully": [...], "not founded": [...] }
      const res = await shareBoard(boardId, usernames);
      const notFound = res['not founded'] || [];
      if (notFound.length) {
        errSlot.appendChild(el(`<div class="banner banner-info">Couldn't find: ${escapeHtml(notFound.join(', '))}</div>`));
      }
      const found = res['founded seccusfully'] || [];
      if (found.length && !notFound.length) {
        modalRoot.innerHTML = '';
      }
    } catch (err) {
      errSlot.appendChild(el(`<div class="banner banner-error">${escapeHtml(err.message)}</div>`));
    } finally {
      btn.disabled = false;
    }
  };
}

function shareBoard(boardId, usernames) {
  return api(`/weblog/${boardId}/share`, { method: 'POST', body: { usernames } });
}

/* ---------------------------- new entry modal ---------------------------- */

function openNewEntryModal() {
  modalRoot.innerHTML = '';
  const overlay = el(`
    <div class="modal-overlay">
      <div class="modal-card">
        <h2 class="modal-title">New entry</h2>
        <div id="modal-error"></div>
        <form id="entry-form">
          <div class="field">
            <label for="e-title">Title</label>
            <input type="text" id="e-title" required>
          </div>
          <div class="field">
            <label for="e-content">Content</label>
            <textarea id="e-content" required></textarea>
          </div>
          <div class="field">
            <label for="e-image">Image (optional)</label>
            <input type="file" id="e-image" accept="image/png,image/jpeg,image/gif,image/webp">
            <span class="field-hint" id="e-image-status">JPG, PNG, GIF or WEBP, up to 5MB.</span>
          </div>
          <div class="field">
            <div class="toggle-row">
              <span class="toggle-label">Privacy</span>
              <label class="switch">
                <input type="checkbox" id="e-private">
                <span class="track"></span>
              </label>
              <span class="privacy-state" id="privacy-label">Public — anyone can read this</span>
            </div>
          </div>
          <div class="field hidden" id="shared-field">
            <label for="e-shared">Share with (usernames)</label>
            <input type="text" id="e-shared" placeholder="alice, bob">
            <span class="field-hint">Comma-separated. Only you and these people can view this entry.</span>
          </div>
          <div class="modal-actions">
            <button type="button" class="btn btn-ghost" id="btn-cancel">Cancel</button>
            <button type="submit" class="btn btn-brass" id="btn-publish">Publish</button>
          </div>
        </form>
      </div>
    </div>
  `);
  modalRoot.appendChild(overlay);

  const privateToggle = overlay.querySelector('#e-private');
  const privacyLabel = overlay.querySelector('#privacy-label');
  const sharedField = overlay.querySelector('#shared-field');
  privateToggle.onchange = () => {
    const isPrivate = privateToggle.checked;
    privacyLabel.textContent = isPrivate
      ? 'Private — only you and people you share with'
      : 'Public — anyone can read this';
    privacyLabel.classList.toggle('is-private', isPrivate);
    sharedField.classList.toggle('hidden', !isPrivate);
  };

  overlay.querySelector('#btn-cancel').onclick = () => modalRoot.innerHTML = '';
  overlay.onclick = (e) => { if (e.target === overlay) modalRoot.innerHTML = ''; };

  overlay.querySelector('#entry-form').onsubmit = async (e) => {
    e.preventDefault();
    const btn = overlay.querySelector('#btn-publish');
    const errSlot = overlay.querySelector('#modal-error');
    btn.disabled = true;
    errSlot.innerHTML = '';

    const isPrivate = privateToggle.checked;
    const imageFile = overlay.querySelector('#e-image').files[0];
    const imageStatus = overlay.querySelector('#e-image-status');

    try {
      let imgPath = '';
      if (imageFile) {
        imageStatus.textContent = 'Uploading image…';
        imgPath = await uploadImage(imageFile);
        imageStatus.textContent = 'Image uploaded.';
      }

      const payload = {
        title: overlay.querySelector('#e-title').value.trim(),
        content: overlay.querySelector('#e-content').value.trim(),
        is_private: isPrivate,
        img_path: imgPath,
      };

      const board = await api('/weblog', { method: 'POST', body: payload });

      if (isPrivate) {
        const usernames = overlay.querySelector('#e-shared').value
          .split(',').map(s => s.trim()).filter(Boolean);
        if (usernames.length) {
          // Sharing is a separate call against the board we just created.
          await shareBoard(board.id, usernames).catch(() => { /* best effort here; can retry from the entry page */ });
        }
      }

      modalRoot.innerHTML = '';
      await loadFeed();
      navigate({ name: 'feed' });
    } catch (err) {
      errSlot.appendChild(el(`<div class="banner banner-error">${escapeHtml(err.message)}</div>`));
    } finally {
      btn.disabled = false;
    }
  };
}

/* ---------------------------- auth actions ---------------------------- */

function logout() {
  // No /logout route on the backend — this is purely a client-side
  // token drop.
  AUTH_TOKEN = null;
  state.user = null;
  state.boards = [];
  clearStoredAuth();
  renderHeader();
  renderAuth();
}

/* ---------------------------- routing ---------------------------- */

function navigate(route) {
  state.route = route;
  renderRoute();
}

function renderRoute() {
  if (!state.user) { renderAuth(); return; }
  if (state.route.name === 'detail') {
    renderDetail(state.route.id);
  } else {
    renderFeed();
  }
}

/* ---------------------------- boot ---------------------------- */

(async function boot() {
  renderHeader();
  if (state.user) {
    // There's no GET /me to validate the stored token up front, so we
    // optimistically trust it and let loadFeed's 401 handling clear it
    // if it turns out to be stale/invalid.
    await loadFeed();
  } else {
    state.loading = false;
  }
  renderRoute();
})();