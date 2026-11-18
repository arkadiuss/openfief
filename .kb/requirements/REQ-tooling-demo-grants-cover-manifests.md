---
title: The demo manifest check must fail when, in the live demo's build, a PluginGrant does not approve a Secret named <instance>--<role> for a role its Plugin's manifest asks for, optional roles included.
status: open
semantic_text: The demo manifest check must fail when, in the live demo's build, a PluginGrant does not approve a Secret named <instance>--<role> for a role its Plugin's manifest asks for, optional roles included.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3890aadaffff8f531f85d7030a1cf3d71a17fedde1eb37f485a641bf2541408b
semantic_inventory:
  - claim_key: CLAIM-CF92676B5FF7256F
    claim_text: The demo manifest check must fail when, in the live demo's build, a PluginGrant does not approve a Secret named <instance>--<role> for a role its Plugin's manifest asks for, optional roles included
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 197
logic_claims:
  - CLAIM-CF92676B5FF7256F
origin:
  kind: agent
  ref: Orchestrating session, after e2e run 44262665311, 2026-11-18
  recorded_at: '2026-11-18T08:39:47.794Z'
id: REQ-tooling-demo-grants-cover-manifests
type: req
---
The demo manifest check must fail when, in the live demo's build, a PluginGrant does not approve a Secret named <instance>--<role> for a role its Plugin's manifest asks for, optional roles included.

## Context

The operator names an instance's Secret <instance>--<role> and keeps a Plugin AwaitingApproval, never Ready, while its PluginGrant leaves out any Secret its manifest asks for, optional ones included. On 2026-11-18 the demo's PluginGrant openrouter approved only openrouter--api-key while the model-openai manifest also asks for an optional ca-bundle, so all four e2e jobs of CI run 44262665311 timed out after five minutes waiting for the model layer, and the live demo would have kept its model provider awaiting approval. The design session asked, for the CRD enum drift found the same morning, that such drift fail a fast check at commit time instead of e2e; this applies the same rule to the demo's grants. Modeled by Claude Code (the orchestrating session); the author has not reviewed it yet.

## Source

> e2e: installing the stack failed: waiting for the model layer, Plugin openrouter: timed out after 5m0s; last state: Ready=False (AwaitingApproval: the plugin requests permissions its PluginGrant does not approve; see status.pendingApproval); ApprovalRequired=True (PermissionsPending: add to the PluginGrant to approve: secrets: openrouter--ca-bundle)

CI run 44262665311, job e2e ghostfolio (kind).

> One addition so this class of drift can't recur: a unit test [...] That turns an e2e-only failure into a fast unit failure at commit time.

The design session, 2026-11-18, on the CRD enum drift.
