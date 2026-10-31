---
title: The first passkey is created and used on localhost through openfief enroll
status: active
tags:
  - acceptance:p10-8
origin:
  kind: agent
  recorded_at: '2026-10-31T12:35:12.107Z'
id: SCEN-e2e-first-passkey-localhost
type: scenario
---
Given the demo's stack with the operator-deployed UI on a local kind or k3d cluster and no route for the owner's hostname, when openfief enroll runs without an origin, it forwards localhost:8080 to the core and prints the one-time link; a browser with a virtual authenticator opens the link on http://localhost, creates the first passkey and lands signed in with the __Host- session cookie; after signing out it signs in again with the same passkey. No flag weakens the origin check. This is the first half of P-10's acceptance criterion 8 (passkey login on localhost); the second hostname and the charts are not verified end to end yet.