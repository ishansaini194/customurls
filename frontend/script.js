/* ─── Config ──────────────────────────────────────────────────────── */
const BACKEND_URL = window.location.origin;
const HISTORY_KEY = 'customurls_history';
const HISTORY_MAX = 20;

/* ─── Utility ─────────────────────────────────────────────────────── */
const $ = id => document.getElementById(id);
const show = el => el.classList.add('show');
const hide = el => el.classList.remove('show');

function toast(msg) {
  const t = $('toast');
  t.querySelector('.toast-text').textContent = msg;
  show(t);
  setTimeout(() => hide(t), 2800);
}

async function copyText(text) {
  try {
    await navigator.clipboard.writeText(text);
    toast('Copied to clipboard');
  } catch {
    toast('Copy failed — select manually');
  }
}

function escHtml(str) {
  return String(str)
    .replace(/&/g, '&amp;').replace(/</g, '&lt;')
    .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function truncate(str, max) {
  return str.length <= max ? str : str.slice(0, max) + '…';
}

// Expiry comes from /shorten as hours; from /stats as an ISO timestamp.
function formatExpiryHours(hours) {
  if (!hours || hours === 0) return 'Never';
  const h = Number(hours);
  if (h >= 8760) return `${Math.round(h / 8760)}y`;
  if (h >= 720) return `${Math.round(h / 720)}mo`;
  if (h >= 24) return `${Math.round(h / 24)}d`;
  return `${h}h`;
}

function formatExpiryDate(iso) {
  if (!iso) return 'Never';
  const then = new Date(iso);
  const diffMs = then - Date.now();
  if (diffMs <= 0) return 'Expired';
  const days = Math.floor(diffMs / 86400000);
  const hrs = Math.floor((diffMs % 86400000) / 3600000);
  if (days > 0) return `${days}d ${hrs}h`;
  return `${hrs}h`;
}

// Extract the short ID (last path segment) from any input form.
function extractShortID(raw) {
  let s = raw.trim();
  try {
    const parsed = new URL(s);
    s = parsed.pathname;
  } catch { /* not a URL — treat as raw */ }
  s = s.replace(/^\/+|\/+$/g, '');
  if (!s) return '';
  const parts = s.split('/');
  return parts[parts.length - 1];
}

// Build the backend QR endpoint URL for a given short URL string.
// The QR endpoint takes the short ID (last path segment).
function qrEndpoint(shortUrl, size) {
  const id = extractShortID(shortUrl);
  const sz = size ? `?size=${size}` : '';
  return `${BACKEND_URL}/qr/${encodeURIComponent(id)}${sz}`;
}

/* ─── Tabs ────────────────────────────────────────────────────────── */
function initTabs() {
  const tabs = document.querySelectorAll('.tab-btn');
  const panels = document.querySelectorAll('.tab-panel');

  tabs.forEach(tab => {
    tab.addEventListener('click', () => {
      tabs.forEach(t => { t.classList.remove('active'); t.setAttribute('aria-selected', 'false'); });
      panels.forEach(p => p.classList.remove('active'));
      tab.classList.add('active');
      tab.setAttribute('aria-selected', 'true');
      $(tab.dataset.panel).classList.add('active');
    });
  });
}

/* ─── Shorten + inline QR ─────────────────────────────────────────── */
function initShorten() {
  const form = $('shorten-form');
  const urlInput = $('url-input');
  const aliasInput = $('alias-input');
  const expiryInput = $('expiry-input');
  const btnText = $('shorten-btn-text');
  const spinner = $('shorten-spinner');
  const errorEl = $('shorten-error');
  const resultCard = $('shorten-result');
  const resultUrl = $('result-url');
  const copyBtn = $('copy-result-btn');
  const openBtn = $('open-result-btn');
  const qrBtn = $('qr-btn');
  const qrOutput = $('qr-output');
  const qrImage = $('qr-image');
  const qrDownload = $('qr-download-btn');

  let currentShortUrl = '';

  form.addEventListener('submit', async e => {
    e.preventDefault();
    hide(errorEl);
    hide(resultCard);
    hide(qrOutput);
    urlInput.classList.remove('error');

    const url = urlInput.value.trim();
    const alias = aliasInput.value.trim();
    const expiry = expiryInput.value.trim();

    if (!url) {
      urlInput.classList.add('error');
      errorEl.textContent = 'Please enter a URL.';
      show(errorEl);
      return;
    }

    try { new URL(url); } catch {
      urlInput.classList.add('error');
      errorEl.textContent = 'Please enter a valid URL (include https://).';
      show(errorEl);
      return;
    }

    btnText.textContent = 'Shortening…';
    spinner.classList.add('show');
    form.querySelector('button[type=submit]').disabled = true;

    try {
      const body = { url };
      if (alias) body.alias = alias;
      if (expiry) body.expiry = parseInt(expiry) * 24; // days → hours

      const res = await fetch(`${BACKEND_URL}/shorten`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      const data = await res.json();

      if (!res.ok) {
        errorEl.textContent = data.error || 'Something went wrong. Try again.';
        show(errorEl);
        return;
      }

      currentShortUrl = data.short;
      resultUrl.textContent = data.short;
      show(resultCard);

      // QR button is now usable
      qrBtn.disabled = false;
      qrBtn.title = 'Generate QR code';

      // Save to history
      addToHistory({
        short: data.short,
        original: data.url || url,
        expiry: data.expiry,            // hours
        clicks: 0,
        savedAt: Date.now(),
      });

      resultCard.scrollIntoView({ behavior: 'smooth', block: 'nearest' });

    } catch {
      errorEl.textContent = 'Cannot reach the server. Is the backend running?';
      show(errorEl);
    } finally {
      btnText.textContent = 'Shorten URL';
      spinner.classList.remove('show');
      form.querySelector('button[type=submit]').disabled = false;
    }
  });

  // QR button — shows QR for the last shortened link (from the backend /qr endpoint)
  qrBtn.addEventListener('click', () => {
    if (!currentShortUrl) return;
    if (qrOutput.classList.contains('show')) {
      hide(qrOutput);
      return;
    }
    qrImage.src = qrEndpoint(currentShortUrl, 240);
    qrImage.onerror = () => toast('Failed to load QR code');
    qrDownload.href = qrEndpoint(currentShortUrl, 512);
    show(qrOutput);
  });

  copyBtn.addEventListener('click', () => copyText(currentShortUrl));
  openBtn.addEventListener('click', () => {
    if (currentShortUrl) window.open(currentShortUrl, '_blank', 'noopener');
  });
}

/* ─── Stats ───────────────────────────────────────────────────────── */
function initStats() {
  const form = $('stats-form');
  const input = $('stats-input');
  const spinner = $('stats-spinner');
  const btnText = $('stats-btn-text');
  const errorEl = $('stats-error');
  const result = $('stats-result');

  form.addEventListener('submit', async e => {
    e.preventDefault();
    hide(errorEl);
    hide(result);

    const shortID = extractShortID(input.value);
    if (!shortID) {
      errorEl.textContent = 'Enter a short ID or full short URL.';
      show(errorEl);
      return;
    }

    btnText.textContent = 'Looking up…';
    spinner.classList.add('show');
    form.querySelector('button[type=submit]').disabled = true;

    try {
      const res = await fetch(`${BACKEND_URL}/stats/${encodeURIComponent(shortID)}`);
      const data = await res.json();

      if (!res.ok) {
        errorEl.textContent = data.error || 'Short URL not found.';
        show(errorEl);
        return;
      }

      $('stat-hits').textContent = data.hits ?? 0;
      $('stat-expiry').textContent = formatExpiryDate(data.expires_at);
      $('stat-shortid').textContent = data.short_id || shortID;
      $('stat-original').innerHTML = data.original_url
        ? `<a href="${escHtml(data.original_url)}" target="_blank" rel="noopener">${escHtml(truncate(data.original_url, 70))}</a>`
        : '—';

      show(result);

      // If this link is in history, refresh its click count
      updateHistoryClicks(data.short_id || shortID, data.hits ?? 0);

    } catch {
      errorEl.textContent = 'Cannot reach the server. Is the backend running?';
      show(errorEl);
    } finally {
      btnText.textContent = 'Get Stats';
      spinner.classList.remove('show');
      form.querySelector('button[type=submit]').disabled = false;
    }
  });
}

/* ─── History (localStorage) ──────────────────────────────────────── */
function loadHistory() {
  try {
    return JSON.parse(localStorage.getItem(HISTORY_KEY)) || [];
  } catch {
    return [];
  }
}

function saveHistory(list) {
  try {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(list));
  } catch {
    // localStorage full or disabled — fail silently
  }
}

function addToHistory(item) {
  let list = loadHistory();
  // De-dupe by short URL — newest wins, moves to top
  list = list.filter(x => x.short !== item.short);
  list.unshift(item);
  if (list.length > HISTORY_MAX) list = list.slice(0, HISTORY_MAX);
  saveHistory(list);
  renderHistory();
}

function updateHistoryClicks(shortID, clicks) {
  const list = loadHistory();
  let changed = false;
  list.forEach(item => {
    // Compare the last path segment of the stored short URL to shortID exactly.
    if (extractShortID(item.short) === shortID) {
      item.clicks = clicks;
      changed = true;
    }
  });
  if (changed) { saveHistory(list); renderHistory(); }
}

function removeHistoryItem(short) {
  const list = loadHistory().filter(x => x.short !== short);
  saveHistory(list);
  renderHistory();
}

function renderHistory() {
  const list = loadHistory();
  const listEl = $('history-list');
  const emptyEl = $('history-empty');

  listEl.innerHTML = '';

  if (list.length === 0) {
    show(emptyEl);
    return;
  }
  hide(emptyEl);

  list.forEach(item => {
    const row = document.createElement('div');
    row.className = 'history-item';

    // QR thumbnail — loaded from the backend /qr endpoint
    const qrWrap = document.createElement('div');
    qrWrap.className = 'history-qr';
    const qrImg = document.createElement('img');
    qrImg.alt = 'QR';
    qrImg.src = qrEndpoint(item.short, 120);
    qrWrap.appendChild(qrImg);

    // Info block
    const info = document.createElement('div');
    info.className = 'history-info';
    info.innerHTML = `
      <div class="history-short">
        <a href="${escHtml(item.short)}" target="_blank" rel="noopener">${escHtml(item.short)}</a>
      </div>
      <div class="history-original" title="${escHtml(item.original)}">${escHtml(item.original)}</div>
      <div class="history-meta">
        <span class="history-meta-item"><strong>${item.clicks ?? 0}</strong> clicks</span>
        <span class="history-meta-item">expires <strong>${formatExpiryHours(item.expiry)}</strong></span>
      </div>
    `;

    // Actions
    const actions = document.createElement('div');
    actions.className = 'history-actions';

    const copyBtn = document.createElement('button');
    copyBtn.className = 'btn btn-icon';
    copyBtn.title = 'Copy';
    copyBtn.innerHTML = `<svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><rect x="5" y="5" width="9" height="9" rx="1.5"/><path d="M11 5V3.5A1.5 1.5 0 009.5 2h-6A1.5 1.5 0 002 3.5v6A1.5 1.5 0 003.5 11H5"/></svg>`;
    copyBtn.addEventListener('click', () => copyText(item.short));

    const delBtn = document.createElement('button');
    delBtn.className = 'btn btn-icon';
    delBtn.title = 'Remove';
    delBtn.innerHTML = `<svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><path d="M3 4h10M6.5 4V2.5h3V4M5 4l.5 9h5L11 4"/></svg>`;
    delBtn.addEventListener('click', () => removeHistoryItem(item.short));

    actions.appendChild(copyBtn);
    actions.appendChild(delBtn);

    row.appendChild(qrWrap);
    row.appendChild(info);
    row.appendChild(actions);
    listEl.appendChild(row);
  });
}

function initHistory() {
  $('history-clear').addEventListener('click', () => {
    if (loadHistory().length === 0) return;
    saveHistory([]);
    renderHistory();
    toast('History cleared');
  });
  renderHistory();
}

/* ─── CLI install copy button ─────────────────────────────────────── */
function initCLICopy() {
  const btn = $('cli-copy-btn');
  if (!btn) return;
  btn.addEventListener('click', async () => {
    const cmd = $('cli-install-cmd').textContent.trim();
    try {
      await navigator.clipboard.writeText(cmd);
      btn.classList.add('copied');
      toast('Install command copied');
      setTimeout(() => btn.classList.remove('copied'), 1500);
    } catch {
      toast('Copy failed — select manually');
    }
  });
}

/* ─── Boot ────────────────────────────────────────────────────────── */
document.addEventListener('DOMContentLoaded', () => {
  initTabs();
  initShorten();
  initStats();
  initHistory();
  initCLICopy();

  document.querySelectorAll('[data-animate]').forEach((el, i) => {
    el.style.animationDelay = `${i * 80}ms`;
    el.classList.add('anim-in');
  });
});