---
title: Every read procedure of the views must either ask the authorizer about an account or be on an explicit list of household-level reads, which name no account and answer every holder of read access alike, so a new read can't skip both.
status: open
firmness: default
semantic_text: Every read procedure of the views must either ask the authorizer about an account or be on an explicit list of household-level reads, which name no account and answer every holder of read access alike, so a new read can't skip both.
origin:
  kind: agent
  ref: 'ModelPortfolio, PriceChecks, Instrument and InstrumentPrices and Plugins name no account, so they''re household-level reads. Document them as such in docs/concepts/read-api.md, or wherever reads are described. Add a generative test in the pattern of TestEveryOwnerProcedureIsClassifiedForTheAuditLog: every read procedure either asks the authorizer per account or is on an explicit household-level list, so a new read can''t skip both. (the owner''s decision S5, in the items for alpha.2 from the reviews of main 0bfeb62)'
  recorded_at: '2026-11-27T08:31:01.508Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dad3cbd4c6b6222c1aabb2e4f7298fbd3c28c391908db9145c9dbb2cc4eabe6e
semantic_inventory:
  - claim_key: CLAIM-A24019A4A37BB392
    claim_text: Every read procedure of the views must either ask the authorizer about an account or be on an explicit list of household-level reads, which name no account
    role: normative
    status: missing
    span:
      start: 0
      end: 155
    subjects:
      - views.read_authorization_class
  - claim_key: CLAIM-DBC631AA4CE9AE94
    claim_text: answer every holder of read access alike, so a new read can't skip both
    role: normative
    status: missing
    span:
      start: 160
      end: 231
    subjects:
      - views.read_authorization_class
logic_claims:
  - CLAIM-A24019A4A37BB392
  - CLAIM-DBC631AA4CE9AE94
id: REQ-views-read-api-household-reads
type: req
---
Every read procedure of the views must either ask the authorizer about an account or be on an explicit list of household-level reads, which name no account and answer every holder of read access alike, so a new read can't skip both.

## Context

The model portfolio, price checks, an instrument and its prices and the installed plugins name no account, so no grant decides them; the owner decided they stay household-level reads (S5 of the reviews of main 0bfeb62), documented in the read API's page, with a test that keeps a new read from skipping both the authorizer and the list.

## Source

> ModelPortfolio, PriceChecks, Instrument and InstrumentPrices and Plugins name no account, so they're household-level reads. Document them as such in docs/concepts/read-api.md, or wherever reads are described. Add a generative test in the pattern of TestEveryOwnerProcedureIsClassifiedForTheAuditLog: every read procedure either asks the authorizer per account or is on an explicit household-level list, so a new read can't skip both. (the owner's decision S5, in the items for alpha.2 from the reviews of main 0bfeb62)