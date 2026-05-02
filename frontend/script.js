/* ─── Config ──────────────────────────────────────────────────────── */
// In production, set BACKEND_URL to your actual API domain
const BACKEND_URL = window.location.origin;

/* ─── QR Library (inline minimal) ────────────────────────────────── */
// We load qrcode.js from CDN, it's injected in HTML

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

/* ─── Tabs ────────────────────────────────────────────────────────── */
function initTabs() {
  const tabs = document.querySelectorAll('.tab-btn');
  const panels = document.querySelectorAll('.tab-panel');

  tabs.forEach(tab => {
    tab.addEventListener('click', () => {
      tabs.forEach(t => t.classList.remove('active'));
      panels.forEach(p => p.classList.remove('active'));
      tab.classList.add('active');
      $(tab.dataset.panel).classList.add('active');
    });
  });
}

/* ─── Shorten Tab ─────────────────────────────────────────────────── */
function initShorten() {
  const form        = $('shorten-form');
  const urlInput    = $('url-input');
  const aliasInput  = $('alias-input');
  const expiryInput = $('expiry-input');
  const btnText     = $('shorten-btn-text');
  const spinner     = $('shorten-spinner');
  const errorEl     = $('shorten-error');
  const resultCard  = $('shorten-result');
  const resultUrl   = $('result-url');
  const copyBtn     = $('copy-result-btn');
  const openBtn     = $('open-result-btn');
  const qrBtn       = $('goto-qr-btn');

  let currentShortUrl = '';

  form.addEventListener('submit', async e => {
    e.preventDefault();
    hide(errorEl);
    hide(resultCard);
    urlInput.classList.remove('error');

    const url    = urlInput.value.trim();
    const alias  = aliasInput.value.trim();
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

    // UI: loading state
    btnText.textContent = 'Shortening…';
    spinner.classList.add('show');
    form.querySelector('button[type=submit]').disabled = true;

    try {
      const body = { url };
      if (alias) body.alias = alias;
      if (expiry) body.expiry = parseInt(expiry) * 24; // days → hours

      const res  = await fetch(`${BACKEND_URL}/shorten`, {
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

  copyBtn.addEventListener('click', () => copyText(currentShortUrl));

  openBtn.addEventListener('click', () => {
    if (currentShortUrl) window.open(currentShortUrl, '_blank', 'noopener');
  });

  qrBtn.addEventListener('click', () => {
    if (!currentShortUrl) return;
    // Switch to QR tab and pre-fill
    document.querySelector('[data-panel="panel-qr"]').click();
    $('qr-url-input').value = currentShortUrl;
    generateQR(currentShortUrl);
  });
}

/* ─── QR Tab ──────────────────────────────────────────────────────── */
function initQR() {
  const form      = $('qr-form');
  const input     = $('qr-url-input');
  const output    = $('qr-output');
  const canvas    = $('qr-canvas');
  const dlBtn     = $('qr-download-btn');
  const copyImgBtn = $('qr-copy-img-btn');
  const errorEl   = $('qr-error');

  form.addEventListener('submit', e => {
    e.preventDefault();
    hide(errorEl);
    const url = input.value.trim();
    if (!url) {
      errorEl.textContent = 'Enter a URL to generate a QR code.';
      show(errorEl);
      return;
    }
    generateQR(url);
  });

  dlBtn.addEventListener('click', () => {
    const link = document.createElement('a');
    link.download = 'customurl-qr.png';
    link.href = canvas.toDataURL('image/png');
    link.click();
  });

  copyImgBtn.addEventListener('click', async () => {
    canvas.toBlob(async blob => {
      try {
        await navigator.clipboard.write([new ClipboardItem({ 'image/png': blob })]);
        toast('QR image copied');
      } catch {
        toast('Copy failed — use Download instead');
      }
    });
  });
}

function generateQR(url) {
  const output  = $('qr-output');
  const canvas  = $('qr-canvas');
  const errorEl = $('qr-error');

  hide(errorEl);

  if (typeof QRCode === 'undefined') {
    errorEl.textContent = 'QR library not loaded. Check your connection.';
    show(errorEl);
    return;
  }

  try {
    QRCode.toCanvas(canvas, url, {
      width: 200,
      margin: 2,
      color: {
        dark:  '#1A1714',
        light: '#FFFFFF',
      },
      errorCorrectionLevel: 'M',
    }, err => {
      if (err) {
        errorEl.textContent = 'Failed to generate QR: ' + err.message;
        show(errorEl);
        return;
      }
      show(output);
    });
  } catch (err) {
    errorEl.textContent = 'QR generation failed.';
    show(errorEl);
  }
}

/* ─── Stats Tab ───────────────────────────────────────────────────── */
function initStats() {
  const form      = $('stats-form');
  const input     = $('stats-input');
  const spinner   = $('stats-spinner');
  const btnText   = $('stats-btn-text');
  const errorEl   = $('stats-error');
  const result    = $('stats-result');

  form.addEventListener('submit', async e => {
    e.preventDefault();
    hide(errorEl);
    hide(result);

    const raw = input.value.trim();
    if (!raw) {
      errorEl.textContent = 'Enter a short ID or full short URL.';
      show(errorEl);
      return;
    }

    // Extract the short ID — handle both "abc123" and "https://customurls.in/abc123"
    let shortID = raw;
    try {
      const parsed = new URL(raw);
      shortID = parsed.pathname.replace(/^\//, '');
    } catch { /* not a URL, treat as raw ID */ }

    if (!shortID) {
      errorEl.textContent = 'Could not extract a short ID from that input.';
      show(errorEl);
      return;
    }

    btnText.textContent = 'Looking up…';
    spinner.classList.add('show');
    form.querySelector('button[type=submit]').disabled = true;

    try {
      const res  = await fetch(`${BACKEND_URL}/stats/${encodeURIComponent(shortID)}`);
      const data = await res.json();

      if (!res.ok) {
        errorEl.textContent = data.error || 'Short URL not found.';
        show(errorEl);
        return;
      }

      // Populate stats UI
      $('stat-hits').textContent     = data.hits ?? data.clicks ?? '—';
      $('stat-expiry').textContent   = formatExpiry(data.expiry);
      $('stat-original').innerHTML   = data.url
        ? `<a href="${escHtml(data.url)}" target="_blank" rel="noopener">${escHtml(truncate(data.url, 60))}</a>`
        : '—';
      $('stat-shortid').textContent  = shortID;

      show(result);

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

function formatExpiry(hours) {
  if (!hours || hours === 0) return 'Never';
  const h = Number(hours);
  if (h >= 8760) return `${Math.round(h / 8760)}y`;
  if (h >= 720)  return `${Math.round(h / 720)}mo`;
  if (h >= 24)   return `${Math.round(h / 24)}d`;
  return `${h}h`;
}

function escHtml(str) {
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function truncate(str, max) {
  return str.length <= max ? str : str.slice(0, max) + '…';
}

/* ─── Boot ────────────────────────────────────────────────────────── */
document.addEventListener('DOMContentLoaded', () => {
  initTabs();
  initShorten();
  initQR();
  initStats();

  // Animate hero on load
  document.querySelectorAll('[data-animate]').forEach((el, i) => {
    el.style.animationDelay = `${i * 80}ms`;
    el.classList.add('anim-in');
  });
});
