// The overview the installed app keeps for reading offline (D-151,
// docs/concepts/ui.md). Loaded on every page, deferred; nothing inline (CSP).
//
// - The overview page carries, in data-offline-overview, only what may be
//   kept: the total's three values, each account's name, institution,
//   sleeve and three values, the allocation's shares, the date they are
//   valued on and when the session ends. This script encrypts it with
//   AES-GCM under a fresh 256-bit key that WebCrypto generates as not
//   extractable, and stores the ciphertext and the CryptoKey object (never
//   its bytes) in IndexedDB, replacing the copy before it.
// - The offline page, served by the service worker in place of the
//   overview when the network is gone, decrypts that copy and shows it,
//   marked as an offline copy with the time it was kept. A copy whose
//   session has ended is wiped unread.
// - The copy and its key are wiped on sign-out (before the form is sent),
//   on the login page, which the UI shows only without a session, and when
//   htmx is told the session is gone (401). The server's Clear-Site-Data on
//   sign-out and on a rejected session wipes them too, and so does the
//   service worker on sign-out.
'use strict';

(function () {
  // The database's name; pwa.go's offlineDB and the service worker's
  // CONFIG.offlineDb spell it the same way.
  const dbName = 'openfief-offline';
  const storeName = 'overview';
  const keyEntry = 'key';
  const copyEntry = 'copy';
  // The kept overview's shape; offline.go's offlineVersion.
  const version = 1;
  const ivBytes = 12;

  function supported() {
    return window.isSecureContext && 'indexedDB' in window && window.crypto && crypto.subtle;
  }

  function request(req) {
    return new Promise(function (resolve, reject) {
      req.onsuccess = function () { resolve(req.result); };
      req.onerror = function () { reject(req.error); };
    });
  }

  function open() {
    const req = indexedDB.open(dbName, 1);
    req.onupgradeneeded = function () {
      req.result.createObjectStore(storeName);
    };
    return request(req);
  }

  // exists reports whether the database is there, without creating it.
  function exists() {
    if (!indexedDB.databases) {
      return Promise.resolve(true);
    }
    return indexedDB.databases().then(function (list) {
      return list.some(function (db) { return db.name === dbName; });
    });
  }

  // wipe deletes the database: the kept overview and its key.
  function wipe() {
    if (!('indexedDB' in window)) {
      return Promise.resolve();
    }
    return new Promise(function (resolve) {
      const req = indexedDB.deleteDatabase(dbName);
      req.onsuccess = req.onerror = req.onblocked = function () { resolve(); };
    });
  }

  // keep encrypts the overview's JSON text under a new key and stores both.
  function keep(text) {
    const overview = JSON.parse(text);
    if (overview.v !== version) {
      return Promise.resolve();
    }
    const plain = new TextEncoder().encode(JSON.stringify({ keptAt: new Date().toISOString(), overview: overview }));
    const iv = crypto.getRandomValues(new Uint8Array(ivBytes));
    return crypto.subtle.generateKey({ name: 'AES-GCM', length: 256 }, false, ['encrypt', 'decrypt']).then(function (key) {
      return crypto.subtle.encrypt({ name: 'AES-GCM', iv: iv }, key, plain).then(function (sealed) {
        return open().then(function (db) {
          return new Promise(function (resolve, reject) {
            const tx = db.transaction(storeName, 'readwrite');
            tx.objectStore(storeName).put(key, keyEntry);
            tx.objectStore(storeName).put({ iv: iv, data: sealed }, copyEntry);
            tx.oncomplete = function () { db.close(); resolve(); };
            tx.onerror = tx.onabort = function () { db.close(); reject(tx.error); };
          });
        });
      });
    });
  }

  // read returns the kept overview decrypted, or null when there is none,
  // it can't be read, or its session has ended, in which case it is wiped.
  function read() {
    return exists().then(function (there) {
      if (!there) {
        return null;
      }
      return open().then(function (db) {
        const store = db.transaction(storeName, 'readonly').objectStore(storeName);
        return Promise.all([request(store.get(keyEntry)), request(store.get(copyEntry))]).finally(function () {
          db.close();
        });
      }).then(function (entries) {
        const key = entries[0];
        const copy = entries[1];
        if (!key || !copy) {
          return null;
        }
        return crypto.subtle.decrypt({ name: 'AES-GCM', iv: copy.iv }, key, copy.data).then(function (plain) {
          const kept = JSON.parse(new TextDecoder().decode(plain));
          const ends = Date.parse(kept.overview.sessionEnds);
          if (kept.overview.v !== version || !(ends > Date.now())) {
            return wipe().then(function () { return null; });
          }
          return kept;
        });
      });
    }).catch(function () {
      return wipe().then(function () { return null; });
    });
  }

  function fill(root, field, text) {
    const el = root.querySelector('[data-offline-' + field + ']');
    if (el) {
      el.textContent = text;
    }
  }

  function cell(row, text, numeric) {
    const td = row.insertCell();
    td.textContent = text;
    if (numeric) {
      td.className = 'num';
    }
    return td;
  }

  function values(row, v) {
    cell(row, v.gross || 'missing', true);
    cell(row, v.net || 'missing', true);
    cell(row, v.liquidation || 'missing', true);
  }

  // show writes the kept overview into the offline page, as text only.
  function show(root, kept) {
    const o = kept.overview;
    const at = new Date(kept.keptAt);
    fill(root, 'stamp', 'The overview as it was when you last opened it, at ' + at.toLocaleString() +
      ', valued on ' + o.date + '. It may be out of date, and nothing else of your portfolio is on this device.');
    fill(root, 'date', o.date);
    fill(root, 'gross', o.total.gross || 'missing');
    fill(root, 'net', o.total.net || 'missing');
    fill(root, 'liquidation', o.total.liquidation || 'missing');
    const accounts = root.querySelector('[data-offline-accounts]');
    o.accounts.forEach(function (a) {
      const row = accounts.insertRow();
      cell(row, a.name);
      cell(row, a.institution);
      cell(row, a.sleeve);
      if (a.valued) {
        values(row, a.values);
      } else {
        const td = cell(row, 'not valued', true);
        td.colSpan = 3;
        td.classList.add('empty');
      }
    });
    fill(root, 'split', o.allocation.value ? o.allocation.value + ' value by ' + o.allocation.by + ', as shares of the total.' : '');
    fill(root, 'by', o.allocation.by || 'Group');
    const shares = root.querySelector('[data-offline-shares]');
    o.allocation.shares.forEach(function (s) {
      const row = shares.insertRow();
      cell(row, s.key);
      cell(row, s.share, true);
    });
    root.hidden = false;
    // layout.js labels tables that htmx brings in for the phone's cards;
    // these rows arrived the same way, after the page loaded.
    root.dispatchEvent(new CustomEvent('htmx:load', { bubbles: true }));
  }

  // signOut wipes the copy before the sign-out form goes.
  function signOut(event) {
    const form = event.target;
    if (!(form instanceof HTMLFormElement) || form.getAttribute('action') !== '/logout') {
      return;
    }
    event.preventDefault();
    wipe().then(function () { form.submit(); });
  }

  document.addEventListener('submit', signOut, true);
  document.addEventListener('htmx:responseError', function (event) {
    if (event.detail && event.detail.xhr && event.detail.xhr.status === 401) {
      wipe();
    }
  });

  function start() {
    if (document.querySelector('[data-offline-wipe]')) {
      wipe();
      return;
    }
    if (!supported()) {
      return;
    }
    const source = document.querySelector('[data-offline-overview]');
    if (source && source.dataset.offlineOverview) {
      keep(source.dataset.offlineOverview).catch(function () {
        // Without a kept copy the app works the same, only not offline.
      });
      return;
    }
    const copy = document.getElementById('offline-copy');
    if (copy && window.location.pathname === '/') {
      read().then(function (kept) {
        if (kept) {
          show(copy, kept);
        }
      });
    }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', start);
  } else {
    start();
  }
})();
