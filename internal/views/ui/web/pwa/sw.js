// OpenFief's service worker (docs/concepts/ui.md). The UI serves it at /sw.js
// with CONFIG above this line: the version, the cache's name and the exact
// list of static files it may cache (styles, scripts, fonts, icons, the
// manifest and the offline page), none of which holds anything of the
// owner's.
//
// What it never does: cache a page, an htmx fragment, a form's answer or any
// other response. Pages hold the owner's record, so a page always comes from
// the network; offline, a navigation gets the static offline page instead.
// Every other request passes through untouched.
//
// The cache is filled when the worker installs and after a sign-in, emptied
// on sign-out, and a new version (any static file changed) drops the old one.
'use strict';

const assets = new Set(CONFIG.assets);

// precache stores every listed file, bypassing the browser's HTTP cache so
// a new version gets the new files.
function precache() {
  return caches.open(CONFIG.cache).then(function (cache) {
    return cache.addAll(CONFIG.assets.map(function (url) {
      return new Request(url, { cache: 'reload', credentials: 'omit' });
    }));
  });
}

// drop deletes every cache of this origin's worker but keep.
function drop(keep) {
  return caches.keys().then(function (names) {
    return Promise.all(names.filter(function (name) {
      return name !== keep;
    }).map(function (name) {
      return caches.delete(name);
    }));
  });
}

self.addEventListener('install', function (event) {
  self.skipWaiting();
  event.waitUntil(precache());
});

self.addEventListener('activate', function (event) {
  event.waitUntil(drop(CONFIG.cache).then(function () {
    return self.clients.claim();
  }));
});

self.addEventListener('fetch', function (event) {
  const request = event.request;
  const url = new URL(request.url);
  if (url.origin !== self.location.origin) {
    return;
  }
  if (request.method === 'POST') {
    // Sign-out empties the cache and sign-in fills it; the request itself
    // goes to the network untouched.
    if (url.pathname === CONFIG.signOut) {
      event.waitUntil(drop(''));
    } else if (CONFIG.signIn.indexOf(url.pathname) >= 0) {
      event.waitUntil(precache());
    }
    return;
  }
  if (request.method !== 'GET') {
    return;
  }
  if (url.search === '' && assets.has(url.pathname)) {
    // A listed static file: from the cache, else from the network, which
    // is not stored.
    event.respondWith(caches.match(request, { cacheName: CONFIG.cache }).then(function (hit) {
      return hit || fetch(request);
    }));
    return;
  }
  if (request.mode === 'navigate') {
    // A page: always from the network, never stored; offline, the static
    // offline page, if the cache holds it.
    event.respondWith(fetch(request).catch(function () {
      return caches.match(CONFIG.offline, { cacheName: CONFIG.cache }).then(function (page) {
        return page || Response.error();
      });
    }));
  }
});
