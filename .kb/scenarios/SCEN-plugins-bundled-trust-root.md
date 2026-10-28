---
title: Verification needs no network beyond the registry
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:16:56.790Z'
id: SCEN-plugins-bundled-trust-root
type: scenario
---
Given the operator binary, when it starts, then it loads the Sigstore public-good trust root embedded in it and verifies bundles without calling Fulcio, Rekor or a TUF mirror; the root is refreshed only by task trustroot:update, which checks it through TUF.