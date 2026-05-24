const BACKEND_URL = window.location.origin;
const HISTORY_KEY = 'customurls_history';
const HISTORY_MAX = 20;

/* Tweakable defaults (host can rewrite this block in place) */
const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "accent": "amber",
  "density": "comfortable",
  "monoEverywhere": false,
  "showCardChrome": true
}/*EDITMODE-END*/;

/* ─── Utility ─────────────────────────────────────────────────────── */
const $ = id => document.getElementById(id);
const show = el => el && el.classList.add('show');
const hide = el => el && el.classList.remove('show');

function toast(msg, kind = 'ok') {
  const t = $('toast');
  if (!t) return;
  t.querySelector('.toast-text').textContent = msg;
  const dot = t.querySelector('.toast-dot');
  if (dot) {
    dot.style.background = kind === 'err'
      ? 'oklch(0.70 0.18 25)'
      : 'oklch(0.78 0.14 150)';
    dot.style.boxShadow = kind === 'err'
      ? '0 0 0 3px oklch(0.70 0.18 25 / 0.18)'
      : '0 0 0 3px oklch(0.78 0.14 150 / 0.18)';
  }
  show(t);
  clearTimeout(t._timer);
  t._timer = setTimeout(() => hide(t), 2400);
}

async function copyText(text) {
  try {
    await navigator.clipboard.writeText(text);
    toast('Copied to clipboard');
    return true;
  } catch {
    try {
      const ta = document.createElement('textarea');
      ta.value = text;
      ta.style.position = 'fixed'; ta.style.left = '-9999px';
      document.body.appendChild(ta);
      ta.select();
      document.execCommand('copy');
      document.body.removeChild(ta);
      toast('Copied to clipboard');
      return true;
    } catch {
      toast('Copy failed — select manually', 'err');
      return false;
    }
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

const LOCK_ICON = '<svg class="lock-icon" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="11" width="16" height="10" rx="2"/><path d="M8 11V7a4 4 0 018 0v4"/></svg>';

function formatExpiryHours(hours) {
  if (!hours || hours === 0) return 'Never';
  const h = Number(hours);
  if (h >= 8760) return `${Math.round(h / 8760)}y`;
  if (h >= 720) return `${Math.round(h / 720)}mo`;
  if (h >= 24) return `${Math.round(h / 24)}d`;
  return `${h}h`;
}

function formatExpiryDate(iso) {
  if (!iso) return { value: 'Never', label: 'Expires in' };
  const then = new Date(iso);
  const diffMs = then - Date.now();
  if (diffMs <= 0) return { value: 'Expired', label: 'Status' };
  const days = Math.floor(diffMs / 86400000);
  const hrs = Math.floor((diffMs % 86400000) / 3600000);
  return {
    value: days > 0 ? `${days}d ${hrs}h` : `${hrs}h`,
    label: 'Expires in',
  };
}

function extractShortID(raw) {
  let s = (raw || '').trim();
  try {
    const parsed = new URL(s);
    s = parsed.pathname;
  } catch { /* not a URL */ }
  s = s.replace(/^\/+|\/+$/g, '');
  if (!s) return '';
  const parts = s.split('/');
  return parts[parts.length - 1];
}

function looksLikeShortUrl(text) {
  try {
    const u = new URL(text.trim());
    const host = window.location.host;
    if (u.host === host || u.host.endsWith('customurls.in')) {
      const seg = u.pathname.replace(/^\/+|\/+$/g, '');
      return seg && !seg.includes('/');
    }
  } catch { }
  return false;
}

function isValidUrl(value) {
  try { new URL(value); return true; } catch { return false; }
}

/* ─── Tabs ───────────────────────────────────────────────────────── */
function setActiveTab(panelId) {
  document.querySelectorAll('.tab-btn').forEach(t => {
    const isActive = t.dataset.panel === panelId;
    t.classList.toggle('active', isActive);
    t.setAttribute('aria-selected', isActive ? 'true' : 'false');
  });
  document.querySelectorAll('.tab-panel').forEach(p => {
    p.classList.toggle('active', p.id === panelId);
  });
  const mode = $('card-title-mode');
  const map = {
    'panel-shorten': 'shorten',
    'panel-qr': 'qr-code',
    'panel-protected': 'protected',
    'panel-stats': 'stats',
  };
  if (mode) mode.textContent = map[panelId] || 'shorten';
}

function initTabs() {
  document.querySelectorAll('.tab-btn').forEach(tab => {
    tab.addEventListener('click', () => setActiveTab(tab.dataset.panel));
  });
}

/* ─── Shorten ────────────────────────────────────────────────────── */
let currentShortUrl = '';

function initShorten() {
  const form = $('shorten-form');
  const urlInput = $('url-input');
  const aliasInput = $('alias-input');
  const aliasHint = $('alias-hint');
  const expiryInput = $('expiry-input');
  const btnText = $('shorten-btn-text');
  const spinner = $('shorten-spinner');
  const errorEl = $('shorten-error');
  const resultCard = $('shorten-result');
  const resultUrl = $('result-url');
  const copyBtn = $('copy-result-btn');
  const openBtn = $('open-result-btn');
  const quickQrBtn = $('quick-qr-btn');
  const quickQrOut = $('quick-qr-output');
  const quickCanvas = $('quick-qr-canvas');
  const quickQrDl = $('quick-qr-download');

  aliasInput.addEventListener('input', () => {
    const v = aliasInput.value.trim();
    if (!v) {
      aliasHint.textContent = 'Optional';
      aliasInput.classList.remove('error');
      return;
    }
    if (!/^[a-zA-Z0-9_-]+$/.test(v)) {
      aliasHint.innerHTML = '<span style="color:var(--err)">Only letters, numbers, dashes and underscores</span>';
      aliasInput.classList.add('error');
    } else {
      aliasHint.textContent = 'Optional';
      aliasInput.classList.remove('error');
    }
  });

  urlInput.addEventListener('paste', (e) => {
    const text = (e.clipboardData || window.clipboardData).getData('text');
    if (text && looksLikeShortUrl(text)) {
      setTimeout(() => {
        if (!confirm('That looks like a customurls.in short link. View its stats instead?')) return;
        setActiveTab('panel-stats');
        $('stats-input').value = text;
        urlInput.value = '';
        $('stats-input').focus();
      }, 0);
    }
  });

  form.addEventListener('submit', async e => {
    e.preventDefault();
    hide(errorEl);
    hide(resultCard);
    hide(quickQrOut);
    urlInput.classList.remove('error');

    const url = urlInput.value.trim();
    const alias = aliasInput.value.trim();
    const expiry = expiryInput.value.trim();

    if (!url) {
      urlInput.classList.add('error');
      errorEl.textContent = 'Please enter a URL.';
      show(errorEl); urlInput.focus();
      return;
    }
    if (!isValidUrl(url)) {
      urlInput.classList.add('error');
      errorEl.textContent = 'Please enter a valid URL (include https://).';
      show(errorEl);
      return;
    }
    if (alias && !/^[a-zA-Z0-9_-]+$/.test(alias)) {
      aliasInput.classList.add('error');
      errorEl.textContent = 'Custom alias can only contain letters, numbers, dashes and underscores.';
      show(errorEl);
      return;
    }

    btnText.textContent = 'Shortening…';
    spinner.classList.add('show');
    form.querySelector('button[type=submit]').disabled = true;

    try {
      const body = { url };
      if (alias) body.alias = alias;
      if (expiry) body.expiry = parseInt(expiry, 10) * 24;

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
      quickQrBtn.disabled = false;
      quickQrBtn.title = 'Generate QR code';

      addToHistory({
        short: data.short,
        original: data.url || url,
        expiry: data.expiry,
        clicks: 0,
        protected: false,
        savedAt: Date.now(),
      });

      toast('Short link created');
    } catch {
      errorEl.textContent = 'Cannot reach the server. Is the backend running?';
      show(errorEl);
    } finally {
      btnText.textContent = 'Shorten';
      spinner.classList.remove('show');
      form.querySelector('button[type=submit]').disabled = false;
    }
  });

  quickQrBtn.addEventListener('click', async () => {
    if (!currentShortUrl) return;
    if (quickQrOut.classList.contains('show')) { hide(quickQrOut); return; }
    try {
      await QRCode.toCanvas(quickCanvas, currentShortUrl, {
        width: 240, margin: 1,
        color: { dark: '#1a1816', light: '#ffffff' },
      });
      show(quickQrOut);
    } catch (err) {
      toast('Failed to generate QR', 'err');
    }
  });

  quickQrDl.addEventListener('click', () => {
    if (!currentShortUrl) return;
    const url = quickCanvas.toDataURL('image/png');
    const a = document.createElement('a');
    a.href = url;
    a.download = `customurl-qr.png`;
    document.body.appendChild(a); a.click(); document.body.removeChild(a);
  });

  copyBtn.addEventListener('click', () => copyText(currentShortUrl));
  openBtn.addEventListener('click', () => {
    if (currentShortUrl) window.open(currentShortUrl, '_blank', 'noopener');
  });

  window.__customurls_getCurrentShort = () => currentShortUrl;
}

/* ─── QR tab (entirely client-side) ──────────────────────────────── */
function initQRTab() {
  const form = $('qr-form');
  const urlIn = $('qr-url-input');
  const colorIn = $('qr-color');
  const sizeIn = $('qr-size');
  const fmtIn = $('qr-format');
  const errEl = $('qr-error');
  const previewBox = $('qr-preview-box');
  const emptyEl = $('qr-preview-empty');
  const canvas = $('qr-canvas');
  const svgMount = $('qr-svg-mount');
  const dlBtn = $('qr-download-btn');

  let lastResult = null; // { dataUrl, format, fileName }

  function clearPreview() {
    canvas.style.display = 'none';
    svgMount.style.display = 'none';
    svgMount.innerHTML = '';
    previewBox.classList.remove('is-filled');
    emptyEl.style.display = '';
    dlBtn.disabled = true;
    lastResult = null;
  }
  clearPreview();

  async function generate() {
    hide(errEl);
    const url = urlIn.value.trim();
    if (!url) {
      errEl.textContent = 'Enter a URL to encode.';
      show(errEl);
      urlIn.focus();
      return;
    }
    if (!isValidUrl(url)) {
      errEl.textContent = 'Please enter a valid URL (include https://).';
      show(errEl);
      return;
    }

    const color = colorIn.value;
    const size = parseInt(sizeIn.value, 10);
    const format = fmtIn.value;

    const opts = {
      width: size,
      margin: 1,
      color: { dark: color, light: '#ffffff' },
      errorCorrectionLevel: 'M',
    };

    try {
      if (format === 'svg') {
        const svgStr = await QRCode.toString(url, { ...opts, type: 'svg' });
        svgMount.innerHTML = svgStr;
        const svgEl = svgMount.querySelector('svg');
        if (svgEl) {
          svgEl.setAttribute('width', '200');
          svgEl.setAttribute('height', '200');
        }
        canvas.style.display = 'none';
        svgMount.style.display = '';
        emptyEl.style.display = 'none';
        previewBox.classList.add('is-filled');

        const blob = new Blob([svgStr], { type: 'image/svg+xml' });
        lastResult = { dataUrl: URL.createObjectURL(blob), format: 'svg', fileName: 'qr.svg' };
      } else {
        canvas.width = size; canvas.height = size;
        await QRCode.toCanvas(canvas, url, opts);
        svgMount.style.display = 'none';
        canvas.style.display = '';
        emptyEl.style.display = 'none';
        previewBox.classList.add('is-filled');
        lastResult = { dataUrl: canvas.toDataURL('image/png'), format: 'png', fileName: 'qr.png' };
      }
      dlBtn.disabled = false;
      toast('QR generated');
    } catch (err) {
      console.error(err);
      errEl.textContent = 'Could not generate QR — try a shorter URL.';
      show(errEl);
    }
  }

  form.addEventListener('submit', e => { e.preventDefault(); generate(); });

  dlBtn.addEventListener('click', () => {
    if (!lastResult) return;
    const a = document.createElement('a');
    a.href = lastResult.dataUrl;
    a.download = lastResult.fileName;
    document.body.appendChild(a); a.click(); document.body.removeChild(a);
  });

  // Live regen if user changes color/size/format AND there's already a value
  [colorIn, sizeIn, fmtIn].forEach(input => {
    input.addEventListener('change', () => {
      if (urlIn.value.trim() && lastResult) generate();
    });
  });
}

/* ─── Protected tab ──────────────────────────────────────────────── */
function initProtected() {
  const form = $('protected-form');
  const urlIn = $('protected-url');
  const pwIn = $('protected-password');
  const alIn = $('protected-alias');
  const expIn = $('protected-expiry');
  const errEl = $('protected-error');
  const result = $('protected-result');
  const resUrl = $('protected-result-url');
  const resPw = $('protected-result-pw');
  const copyBtn = $('protected-copy-btn');
  const spinner = $('protected-spinner');
  const btnText = $('protected-btn-text');

  let lastShort = '';

  form.addEventListener('submit', async e => {
    e.preventDefault();
    hide(errEl); hide(result);

    const url = urlIn.value.trim();
    const pw = pwIn.value;
    const al = alIn.value.trim();
    const exp = expIn.value.trim();

    if (!url || !isValidUrl(url)) {
      errEl.textContent = 'Please enter a valid URL (include https://).'; show(errEl); return;
    }
    if (!pw) {
      errEl.textContent = 'Password is required for a protected link.'; show(errEl); return;
    }
    if (al && !/^[a-zA-Z0-9_-]+$/.test(al)) {
      errEl.textContent = 'Alias can only contain letters, numbers, dashes and underscores.'; show(errEl); return;
    }

    btnText.textContent = 'Creating…';
    spinner.classList.add('show');
    form.querySelector('button[type=submit]').disabled = true;

    try {
      const body = { url, password: pw };
      if (al) body.alias = al;
      if (exp) body.expiry = parseInt(exp, 10) * 24;

      const res = await fetch(`${BACKEND_URL}/shorten`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      const data = await res.json();

      if (!res.ok) {
        errEl.textContent = data.error || 'Something went wrong. Try again.'; show(errEl); return;
      }

      lastShort = data.short;
      resUrl.textContent = data.short;
      resPw.textContent = pw;
      show(result);

      addToHistory({
        short: data.short,
        original: data.url || url,
        expiry: data.expiry,
        clicks: 0,
        protected: true,
        savedAt: Date.now(),
      });

      pwIn.value = '';
      toast('Protected link created');
    } catch {
      errEl.textContent = 'Cannot reach the server. Is the backend running?'; show(errEl);
    } finally {
      btnText.textContent = 'Create protected link';
      spinner.classList.remove('show');
      form.querySelector('button[type=submit]').disabled = false;
    }
  });

  copyBtn.addEventListener('click', () => copyText(lastShort));
}

/* ─── Stats ─────────────────────────────────────────────────────── */
function initStats() {
  const form = $('stats-form');
  const input = $('stats-input');
  const spinner = $('stats-spinner');
  const btnText = $('stats-btn-text');
  const errorEl = $('stats-error');
  const result = $('stats-result');

  form.addEventListener('submit', async e => {
    e.preventDefault();
    hide(errorEl); hide(result);

    const shortID = extractShortID(input.value);
    if (!shortID) {
      errorEl.textContent = 'Enter a short ID or full short URL.'; show(errorEl); return;
    }

    btnText.textContent = 'Looking up…';
    spinner.classList.add('show');
    form.querySelector('button[type=submit]').disabled = true;

    try {
      const res = await fetch(`${BACKEND_URL}/stats/${encodeURIComponent(shortID)}`);
      const data = await res.json();

      if (!res.ok) {
        errorEl.textContent = data.error || 'Short URL not found.'; show(errorEl); return;
      }

      $('stat-hits').textContent = data.hits ?? 0;
      const exp = formatExpiryDate(data.expires_at);
      $('stat-expiry').textContent = exp.value;
      $('stat-expiry-label').textContent = exp.label;

      $('stat-shortid').textContent = data.short_id || shortID;
      $('stat-original').innerHTML = data.original_url
        ? `<a href="${escHtml(data.original_url)}" target="_blank" rel="noopener">${escHtml(truncate(data.original_url, 70))}</a>`
        : '—';
      $('stat-protected').textContent = data.protected ? 'Yes — password required' : 'No';

      show(result);
      updateHistoryClicks(data.short_id || shortID, data.hits ?? 0);
    } catch {
      errorEl.textContent = 'Cannot reach the server. Is the backend running?'; show(errorEl);
    } finally {
      btnText.textContent = 'Get Stats';
      spinner.classList.remove('show');
      form.querySelector('button[type=submit]').disabled = false;
    }
  });
}

/* ─── History (localStorage) ────────────────────────────────────── */
function loadHistory() {
  try { return JSON.parse(localStorage.getItem(HISTORY_KEY)) || []; }
  catch { return []; }
}
function saveHistory(list) {
  try { localStorage.setItem(HISTORY_KEY, JSON.stringify(list)); } catch { }
}
function addToHistory(item) {
  let list = loadHistory();
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
    if (extractShortID(item.short) === shortID) {
      item.clicks = clicks;
      changed = true;
    }
  });
  if (changed) { saveHistory(list); renderHistory(); }
}
function removeHistoryItem(short) {
  saveHistory(loadHistory().filter(x => x.short !== short));
  renderHistory();
  toast('Link removed from history');
}

let _historyFilter = '';

function renderHistory() {
  const all = loadHistory();
  const list = _historyFilter
    ? all.filter(it =>
      (it.short || '').toLowerCase().includes(_historyFilter) ||
      (it.original || '').toLowerCase().includes(_historyFilter))
    : all;

  const listEl = $('history-list');
  const emptyEl = $('history-empty');
  const countEl = $('history-count');

  countEl.textContent = all.length;
  listEl.innerHTML = '';

  if (list.length === 0) {
    show(emptyEl);
    emptyEl.querySelector('.hint').textContent = all.length === 0
      ? "Shorten a URL above and it'll show up here — stored locally in your browser."
      : `No links match "${_historyFilter}".`;
    emptyEl.firstChild.textContent = all.length === 0 ? 'No links yet' : 'No matches';
    return;
  }
  hide(emptyEl);

  list.forEach(item => {
    const row = document.createElement('div');
    row.className = 'history-item';

    const qrWrap = document.createElement('div');
    qrWrap.className = 'history-qr';
    const qrCanvas = document.createElement('canvas');
    qrCanvas.width = 80; qrCanvas.height = 80;
    qrWrap.appendChild(qrCanvas);
    // generate locally
    QRCode.toCanvas(qrCanvas, item.short, {
      width: 80, margin: 0, color: { dark: '#1a1816', light: '#ffffff' },
    }).catch(() => { qrWrap.style.background = 'var(--surface-2)'; qrCanvas.style.display = 'none'; });

    const lockIcon = item.protected ? LOCK_ICON : '';

    const info = document.createElement('div');
    info.className = 'history-info';
    info.innerHTML = `
      <div class="history-short">
        ${lockIcon}<a href="${escHtml(item.short)}" target="_blank" rel="noopener">${escHtml(item.short)}</a>
      </div>
      <div class="history-original" title="${escHtml(item.original)}">${escHtml(item.original)}</div>
      <div class="history-meta">
        <span class="history-meta-item"><strong>${item.clicks ?? 0}</strong> clicks</span>
        <span class="history-meta-item">expires <strong>${formatExpiryHours(item.expiry)}</strong></span>
        ${item.protected ? '<span class="history-meta-item">password</span>' : ''}
      </div>
    `;

    const actions = document.createElement('div');
    actions.className = 'history-actions';

    const copyBtn = document.createElement('button');
    copyBtn.className = 'btn btn-icon';
    copyBtn.title = 'Copy';
    copyBtn.innerHTML = `<svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><rect x="5" y="5" width="9" height="9" rx="1.5"/><path d="M11 5V3.5A1.5 1.5 0 009.5 2h-6A1.5 1.5 0 002 3.5v6A1.5 1.5 0 003.5 11H5"/></svg>`;
    copyBtn.addEventListener('click', () => copyText(item.short));

    const statsBtn = document.createElement('button');
    statsBtn.className = 'btn btn-icon';
    statsBtn.title = 'Check stats';
    statsBtn.innerHTML = `<svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M2 12l3-4 3 2 3-5 3 3"/></svg>`;
    statsBtn.addEventListener('click', () => {
      setActiveTab('panel-stats');
      $('stats-input').value = item.short;
      $('stats-form').dispatchEvent(new Event('submit'));
      document.getElementById('app').scrollIntoView({ behavior: 'smooth', block: 'start' });
    });

    const delBtn = document.createElement('button');
    delBtn.className = 'btn btn-icon';
    delBtn.title = 'Remove';
    delBtn.innerHTML = `<svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><path d="M3 4h10M6.5 4V2.5h3V4M5 4l.5 9h5L11 4"/></svg>`;
    delBtn.addEventListener('click', () => removeHistoryItem(item.short));

    actions.appendChild(copyBtn);
    actions.appendChild(statsBtn);
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
    if (!confirm('Clear all link history?')) return;
    saveHistory([]); renderHistory(); toast('History cleared');
  });

  $('history-export').addEventListener('click', () => {
    const list = loadHistory();
    if (list.length === 0) { toast('Nothing to export', 'err'); return; }
    const blob = new Blob([JSON.stringify(list, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `customurls-history-${new Date().toISOString().slice(0, 10)}.json`;
    document.body.appendChild(a); a.click(); document.body.removeChild(a);
    URL.revokeObjectURL(url);
    toast(`Exported ${list.length} link${list.length === 1 ? '' : 's'}`);
  });

  $('history-search').addEventListener('input', e => {
    _historyFilter = e.target.value.trim().toLowerCase();
    renderHistory();
  });

  renderHistory();
}

/* ─── CLI install copy ───────────────────────────────────────────── */
function initCLICopy() {
  const btn = $('cli-copy-btn');
  if (!btn) return;
  btn.addEventListener('click', async () => {
    const cmd = $('cli-install-cmd').textContent.trim();
    const ok = await copyText(cmd);
    if (ok) {
      btn.classList.add('copied');
      setTimeout(() => btn.classList.remove('copied'), 1500);
    }
  });
}

/* ─── Keyboard shortcuts ─────────────────────────────────────────── */
function initShortcuts() {
  document.addEventListener('keydown', (e) => {
    const meta = e.metaKey || e.ctrlKey;

    if (meta && e.key.toLowerCase() === 'k') {
      e.preventDefault();
      setActiveTab('panel-shorten');
      $('url-input').focus();
      $('url-input').select();
    }

    if (meta && e.key === 'Enter') {
      const activePanel = document.querySelector('.tab-panel.active');
      if (activePanel) {
        const form = activePanel.querySelector('form');
        if (form) { e.preventDefault(); form.requestSubmit(); }
      }
    }

    if (e.key === 'Escape') {
      document.querySelectorAll('.error-msg.show').forEach(el => hide(el));
    }
  });
}

/* ─── Tweaks panel ──────────────────────────────────────────────── */
const TWEAK_STATE = { ...TWEAK_DEFAULTS };

const ACCENT_HUES = {
  amber: 55,
  rose: 18,
  violet: 285,
  cyan: 210,
  lime: 135,
};

function applyTweaks() {
  const root = document.documentElement;
  const h = ACCENT_HUES[TWEAK_STATE.accent] ?? 55;
  root.style.setProperty('--accent-h', h);

  if (TWEAK_STATE.density === 'compact') {
    root.style.setProperty('--r-lg', '10px');
    root.style.setProperty('--r-md', '8px');
  } else {
    root.style.removeProperty('--r-lg');
    root.style.removeProperty('--r-md');
  }

  if (TWEAK_STATE.monoEverywhere) {
    root.style.setProperty('--font-sans', 'Geist Mono, ui-monospace, monospace');
  } else {
    root.style.removeProperty('--font-sans');
  }

  document.querySelectorAll('.card-chrome').forEach(el => {
    el.style.display = TWEAK_STATE.showCardChrome ? '' : 'none';
  });
}

function persistTweaks() {
  window.parent.postMessage({ type: '__edit_mode_set_keys', edits: TWEAK_STATE }, '*');
}

function buildTweaksPanel() {
  const root = $('tweaks-root');
  if (!root) return;
  root.innerHTML = `
    <div class="tweaks-panel" role="dialog" aria-label="Tweaks">
      <div class="tweaks-header">
        <div class="tweaks-title">Tweaks</div>
        <button class="tweaks-close" aria-label="Close">×</button>
      </div>
      <div class="tweaks-body">
        <div class="tweak-section">
          <div class="tweak-label">Accent hue</div>
          <div class="tweak-swatches" data-tweak="accent">
            ${Object.entries(ACCENT_HUES).map(([name, hue]) => `
              <button class="tweak-swatch ${name === TWEAK_STATE.accent ? 'is-active' : ''}"
                      data-value="${name}" title="${name}"
                      style="background:oklch(0.74 0.16 ${hue})"></button>
            `).join('')}
          </div>
        </div>
        <div class="tweak-section">
          <div class="tweak-label">Density</div>
          <div class="tweak-radio" data-tweak="density">
            <button data-value="comfortable" class="${TWEAK_STATE.density === 'comfortable' ? 'is-active' : ''}">Comfortable</button>
            <button data-value="compact" class="${TWEAK_STATE.density === 'compact' ? 'is-active' : ''}">Compact</button>
          </div>
        </div>
        <div class="tweak-section">
          <label class="tweak-toggle">
            <input type="checkbox" data-tweak="monoEverywhere" ${TWEAK_STATE.monoEverywhere ? 'checked' : ''} />
            <span>Mono everywhere</span>
          </label>
          <label class="tweak-toggle">
            <input type="checkbox" data-tweak="showCardChrome" ${TWEAK_STATE.showCardChrome ? 'checked' : ''} />
            <span>Terminal chrome bar</span>
          </label>
        </div>
        <div class="tweak-footer">
          <span class="tweak-kbd">⌘K</span> focus &nbsp;·&nbsp;
          <span class="tweak-kbd">⌘↵</span> submit
        </div>
      </div>
    </div>
  `;

  root.querySelectorAll('.tweak-swatch').forEach(btn => {
    btn.addEventListener('click', () => {
      TWEAK_STATE.accent = btn.dataset.value;
      root.querySelectorAll('[data-tweak="accent"] .tweak-swatch')
        .forEach(b => b.classList.toggle('is-active', b === btn));
      applyTweaks(); persistTweaks();
    });
  });

  root.querySelectorAll('[data-tweak="density"] button').forEach(btn => {
    btn.addEventListener('click', () => {
      TWEAK_STATE.density = btn.dataset.value;
      root.querySelectorAll('[data-tweak="density"] button')
        .forEach(b => b.classList.toggle('is-active', b === btn));
      applyTweaks(); persistTweaks();
    });
  });

  root.querySelectorAll('[type="checkbox"][data-tweak]').forEach(cb => {
    cb.addEventListener('change', () => {
      TWEAK_STATE[cb.dataset.tweak] = cb.checked;
      applyTweaks(); persistTweaks();
    });
  });

  root.querySelector('.tweaks-close').addEventListener('click', () => {
    root.classList.remove('open');
    window.parent.postMessage({ type: '__edit_mode_dismissed' }, '*');
  });
}

function initTweaksHost() {
  window.addEventListener('message', (e) => {
    const d = e.data || {};
    if (d.type === '__activate_edit_mode') {
      $('tweaks-root').classList.add('open');
    } else if (d.type === '__deactivate_edit_mode') {
      $('tweaks-root').classList.remove('open');
    }
  });
  buildTweaksPanel();
  applyTweaks();
  window.parent.postMessage({ type: '__edit_mode_available' }, '*');
}

/* ─── Boot ──────────────────────────────────────────────────────── */
document.addEventListener('DOMContentLoaded', () => {
  document.documentElement.classList.add('js-loaded');
  initTabs();
  initShorten();
  initQRTab();
  initProtected();
  initStats();
  initHistory();
  initCLICopy();
  initShortcuts();
  initTweaksHost();

  document.querySelectorAll('[data-animate]').forEach((el, i) => {
    el.style.animationDelay = `${i * 90}ms`;
    el.classList.add('anim-in');
  });
});
