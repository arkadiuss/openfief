---
title: task demo:model-secret and task demo:turnstile-secret create or replace the demo's Secret (openrouter--api-key, ui--turnstile-secret) by reading the value locally without echo and sending it over SSH on standard input into kubectl create secret --dry-run=client | kubectl apply, so the value is in no argument, file, git, cloud-init or CI and is printed nowhere; they refuse without DEMO_SSH=keys, without a value or with a value that has white space in it, and running one again rotates the Secret.
status: open
firmness: default
semantic_text: task demo:model-secret and task demo:turnstile-secret create or replace the demo's Secret (openrouter--api-key, ui--turnstile-secret) by reading the value locally without echo and sending it over SSH on standard input into kubectl create secret --dry-run=client | kubectl apply, so the value is in no argument, file, git, cloud-init or CI and is printed nowhere; they refuse without DEMO_SSH=keys, without a value or with a value that has white space in it, and running one again rotates the Secret.
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T19:05:43.689Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 67f33c04b08132f375d9f038a687d357749d158c4259ab8e508e1059d8f2932b
semantic_inventory:
  - claim_key: CLAIM-C0A4A4770369E285
    claim_text: task demo:model-secret and task demo:turnstile-secret create or replace the demo's Secret (openrouter--api-key, ui--turnstile-secret) by reading the value locally without echo and sending it over SSH on standard input into kubectl create secret --dry-run=client | kubectl apply, so the value is in no argument, file, git, cloud-init or CI and is printed nowhere
    role: descriptive
    status: missing
    span:
      start: 0
      end: 361
  - claim_key: CLAIM-5B3F0D28FC6CC542
    claim_text: they refuse without DEMO_SSH=keys, without a value or with a value that has white space in it, and running one again rotates the Secret
    role: descriptive
    status: missing
    span:
      start: 363
      end: 498
logic_claims:
  - CLAIM-C0A4A4770369E285
  - CLAIM-5B3F0D28FC6CC542
id: REQ-tooling-demo-secret-over-ssh
type: req
---
task demo:model-secret and task demo:turnstile-secret create or replace the demo's Secret (openrouter--api-key, ui--turnstile-secret) by reading the value locally without echo and sending it over SSH on standard input into kubectl create secret --dry-run=client | kubectl apply, so the value is in no argument, file, git, cloud-init or CI and is printed nowhere; they refuse without DEMO_SSH=keys, without a value or with a value that has white space in it, and running one again rotates the Secret.

## Context

The model's key and the Turnstile secret key must reach the cluster without being in git (D-29), in cloud-init or in a CI job that would then need an SSH key to the machine. An out-of-band step the owner runs, with SSH open only while it runs, keeps the value in the owner's shell and on the wire. Create-or-replace makes rotation the same command.

## Source

> The design session's M5 corrections: task demo:model-secret reads the key locally with read -rs, pipes it over SSH into k3s kubectl -n openfief create secret generic openrouter--api-key --from-file=api-key=/dev/stdin --dry-run=client -o yaml | kubectl apply -f - (create or replace), needs DEMO_SSH=keys while it runs; no secret in git or cloud-init, no SSH key in CI, nothing printed; rotation is re-running it; a script test with a fake ssh (M5 brief, 2026-11-14).