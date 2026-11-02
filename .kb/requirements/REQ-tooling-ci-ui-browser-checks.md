---
title: CI must run the built-in UI's browser checks on every push to main and on a pull request that changes what they read, and a release must not build or publish anything until they pass.
status: open
text_ref: docs/dev/testing.md#browser-checks
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-testing-md-browser-checks
  - lane:strict
semantic_text: CI must run the built-in UI's browser checks on every push to main and on a pull request that changes what they read, and a release must not build or publish anything until they pass.
logic_claims:
  - CLAIM-7BA66D6B3453A53E
semantic_clauses:
  - CI must run the built-in UI's browser checks on every push to main and on a pull request that changes what they read, and a release must not build or publish anything until they pass.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7cc0e510ba26e187b12d4e2160df91b2eb596005bfcf261aa5a14df47605c688
semantic_inventory:
  - claim_key: CLAIM-7BA66D6B3453A53E
    claim_text: CI must run the built-in UI's browser checks on every push to main and on a pull request that changes what they read, and a release must not build or publish anything until they pass
    role: normative
    status: modeled
    span:
      start: 0
      end: 182
origin:
  kind: agent
  recorded_at: '2026-11-02T12:07:11.672Z'
id: REQ-tooling-ci-ui-browser-checks
type: req
---
CI must run the built-in UI's browser checks on every push to main and on a pull request that changes what they read, and a release must not build or publish anything until they pass.

## Context

task ui:browser found that /plugins and the settings page scrolled 751 px sideways on a phone, but nothing ran it, so the regression reached the alpha.1 release head. It stays out of task check, which must not need Chromium; CI runs it as its own job (.github/workflows/ui-browser.yml) when tools/affected says a pull request can affect the UI, on every push to main, and release.yml runs it before the plugin images, which every other release job follows. Modeled by Claude Code for the phone sideways-scroll fix on the alpha.1 release head; the author decided that ui:browser runs in CI and gates a release, and has not reviewed this wording yet.

## Source

docs/dev/testing.md, Browser checks (the design session's rule): ui:browser runs in CI on every change that can affect the UI, and a release needs it green.
