---
title: The owner's first run works on a bundle install with the CLI built from the checkout
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T21:52:04.274Z'
id: SCEN-e2e-bundle-first-run
type: scenario
---
Given OpenFief installed from the release bundle with its own database on a local k3d or kind cluster, and the openfief CLI built from the same checkout, when the owner runs `openfief plugin install ui` and answers yes, then the operator runs the built-in UI with its browser path; and when the owner runs `openfief enroll` and opens its link, then the browser creates the first passkey through the core, signs out and signs in again, and Ctrl-C stops the CLI cleanly. Only downloading the CLI stays outside the suite.