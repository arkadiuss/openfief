---
title: Every response carries a strict CSP and nothing is inline
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:03.064Z'
id: SCEN-ui-strict-csp
type: scenario
---
Given any page or static file, when the UI serves it, then it carries a Content-Security-Policy allowing scripts and styles from the UI itself only and no framing, plus nosniff, no-referrer and HSTS on HTTPS; no template holds an inline script, style or handler, and the vendored libraries are the pinned releases served from the UI.