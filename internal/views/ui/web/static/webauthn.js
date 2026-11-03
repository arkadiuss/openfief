// Passkey ceremonies for OpenFief's built-in UI (D-26, D-60). The UI's server
// relays each step to the core's AuthService; options and credentials travel
// in the WebAuthn Level 3 JSON forms, so this file only parses and serializes
// them with the browser's own helpers. No library, nothing inline (CSP).
'use strict';

(function () {
  // The enrollment code arrives in the link's fragment (#code=...), which the
  // browser never sends to a server. Read it once and drop it from the URL
  // and the history, so it doesn't linger in the address bar.
  const params = new URLSearchParams(window.location.hash.slice(1));
  const enrollmentCode = params.get('code') || '';
  if (enrollmentCode) {
    history.replaceState(null, '', window.location.pathname);
  }

  function supported() {
    return typeof window.PublicKeyCredential === 'function' &&
      typeof PublicKeyCredential.parseRequestOptionsFromJSON === 'function' &&
      typeof PublicKeyCredential.parseCreationOptionsFromJSON === 'function';
  }

  async function postJSON(url, body) {
    const response = await fetch(url, {
      method: 'POST',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    let data = {};
    try {
      data = await response.json();
    } catch (_) {
      data = {};
    }
    if (!response.ok) {
      const err = new Error(data.error || 'The request failed (' + response.status + ').');
      // The core wants a passkey check before this action (D-183).
      err.passkeyCheck = data.passkeyCheck === true;
      throw err;
    }
    return data;
  }

  // check confirms with a passkey in the signed-in session (D-183): an
  // assertion like a login's, after which enrolling or removing a passkey,
  // issuing an enrollment link and creating a script token are allowed for
  // five minutes.
  async function check() {
    const begin = await postJSON('/passkey-check/begin', {});
    const publicKey = PublicKeyCredential.parseRequestOptionsFromJSON(JSON.parse(begin.options));
    const credential = await navigator.credentials.get({ publicKey });
    await postJSON('/passkey-check/finish', {
      ceremonyId: begin.ceremonyId,
      credential: JSON.stringify(credential.toJSON()),
    });
  }

  async function login() {
    const begin = await postJSON('/login/passkey/begin', {});
    const publicKey = PublicKeyCredential.parseRequestOptionsFromJSON(JSON.parse(begin.options));
    const credential = await navigator.credentials.get({ publicKey });
    const done = await postJSON('/login/passkey/finish', {
      ceremonyId: begin.ceremonyId,
      credential: JSON.stringify(credential.toJSON()),
    });
    window.location.assign(done.redirect);
  }

  async function register() {
    const grant = enrollmentCode ? { code: enrollmentCode } : {};
    let begin;
    try {
      begin = await postJSON('/enroll/begin', grant);
    } catch (err) {
      if (!err.passkeyCheck) {
        throw err;
      }
      await check();
      begin = await postJSON('/enroll/begin', grant);
    }
    const publicKey = PublicKeyCredential.parseCreationOptionsFromJSON(JSON.parse(begin.options));
    const credential = await navigator.credentials.create({ publicKey });
    const nameInput = document.getElementById('passkey-name');
    const done = await postJSON('/enroll/finish', {
      ceremonyId: begin.ceremonyId,
      credential: JSON.stringify(credential.toJSON()),
      name: nameInput ? nameInput.value : '',
    });
    window.location.assign(done.redirect);
  }

  // retry confirms with a passkey, then posts the action the core refused
  // for want of a passkey check again, as it was.
  async function retry(button) {
    await check();
    document.getElementById(button.dataset.retry).submit();
  }

  function run(button) {
    switch (button.dataset.webauthn) {
      case 'login':
        return login();
      case 'check':
        return retry(button);
      default:
        return register();
    }
  }

  function showError(button, message) {
    const scope = button.closest('section, form') || button.parentElement;
    const box = scope.querySelector('[data-webauthn-error]');
    if (box) {
      box.textContent = message;
      box.hidden = false;
    }
  }

  document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('[data-webauthn]').forEach(function (button) {
      button.addEventListener('click', async function () {
        if (!supported()) {
          showError(button, 'This browser lacks the WebAuthn JSON helpers passkeys need here. Use a current browser.');
          return;
        }
        button.disabled = true;
        try {
          await run(button);
        } catch (err) {
          const cancelled = err && err.name === 'NotAllowedError';
          showError(button, cancelled ? 'The passkey prompt was cancelled or timed out.' : err.message);
        } finally {
          button.disabled = false;
        }
      });
    });
  });
})();
