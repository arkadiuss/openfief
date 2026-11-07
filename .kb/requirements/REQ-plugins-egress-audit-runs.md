---
title: 'After each demo deploy and after the release smoke test''s install, the egress proxy''s log must be audited against the cluster''s PluginGrants: on the demo by a Job at the egress proxy''s verified digest, re-created by every deploy, with a Role that only lists pods and grants and reads pod logs.'
status: open
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: 'After each demo deploy and after the release smoke test''s install, the egress proxy''s log must be audited against the cluster''s PluginGrants: on the demo by a Job at the egress proxy''s verified digest, re-created by every deploy, with a Role that only lists pods and grants and reads pod logs.'
logic_claims:
  - CLAIM-179A7627DD27815A
semantic_clauses:
  - 'After each demo deploy and after the release smoke test''s install, the egress proxy''s log must be audited against the cluster''s PluginGrants: on the demo by a Job at the egress proxy''s verified digest, re-created by every deploy, with a Role that only lists pods and grants and reads pod logs.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f08ebff794c791711fbcbebb3fa6d0dc662ca4c8f94748fb6ad81068289f69d4
semantic_inventory:
  - claim_key: CLAIM-179A7627DD27815A
    claim_text: 'After each demo deploy and after the release smoke test''s install, the egress proxy''s log must be audited against the cluster''s PluginGrants: on the demo by a Job at the egress proxy''s verified digest, re-created by every deploy, with a Role that only lists pods and grants and reads pod logs'
    role: normative
    status: modeled
    span:
      start: 0
      end: 292
origin:
  kind: agent
  recorded_at: '2026-11-07T03:31:07.371Z'
id: REQ-plugins-egress-audit-runs
type: req
---
After each demo deploy and after the release smoke test's install, the egress proxy's log must be audited against the cluster's PluginGrants: on the demo by a Job at the egress proxy's verified digest, re-created by every deploy, with a Role that only lists pods and grants and reads pod logs.

## Context

The demo's nbp instance reaches NBP's real API, so the demo is where real allowed traffic is audited; the release smoke test audits a published release's install. The deploy job holds no cluster credential (GitOps pull), so the demo's result is the Flux layer's state on the machine. Modeled by Claude Code for the D-200 follow-ups; the author has not reviewed the wording.

## Source

> **Egress.** Outbound traffic goes only to configured destinations (IBKR's Flex service, NBP, GUS, the market-data vendor, the user's Ghostfolio), verified from the egress proxy's logs.

design-decisions:D-200#16 (docs/DESIGN.md, D-200, criterion 16); GR-12
