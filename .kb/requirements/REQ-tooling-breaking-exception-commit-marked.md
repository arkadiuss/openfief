---
title: A commit that adds a breaking-change exception must be marked breaking, with `!` or a BREAKING CHANGE footer, or the commit-msg hook and proto:breaking refuse it.
status: open
text_ref: docs/dev/pipeline.md#breaking-contract-changes
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-pipeline-md-breaking-contract-changes
  - lane:strict
semantic_text: A commit that adds a breaking-change exception must be marked breaking, with `!` or a BREAKING CHANGE footer, or the commit-msg hook and proto:breaking refuse it.
logic_claims:
  - CLAIM-0B50F30D5BCAF230
semantic_clauses:
  - A commit that adds a breaking-change exception must be marked breaking, with `!` or a BREAKING CHANGE footer, or the commit-msg hook and proto:breaking refuse it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 22346c0be3fe87f8f6bd3e51ef08dc375cbfa976292ed34bd5d8a7988a8d742e
semantic_inventory:
  - claim_key: CLAIM-0B50F30D5BCAF230
    claim_text: A commit that adds a breaking-change exception must be marked breaking, with `!` or a BREAKING CHANGE footer, or the commit-msg hook and proto:breaking refuse it
    role: normative
    status: modeled
    span:
      start: 0
      end: 161
origin:
  kind: agent
  recorded_at: '2026-11-01T19:21:43.479Z'
id: REQ-tooling-breaking-exception-commit-marked
type: req
---
A commit that adds a breaking-change exception must be marked breaking, with `!` or a BREAKING CHANGE footer, or the commit-msg hook and proto:breaking refuse it.

## Context

D-157 removes the read API's ListBrokerAccounts as the one breaking change the breaking-change check lets through; docs/dev/pipeline.md says the commit carries `!` or a BREAKING CHANGE footer so the release notes list it. This refines REQ-tooling-breaking-check-one-exception with how the commit is checked. Asked by the author through D-157 and the pipeline rule.

## Source

> A breaking contract change goes in only when an author decision names it [...] The commit marks it breaking (`!` or a `BREAKING CHANGE:` footer), so the release notes list it.

docs/dev/pipeline.md
