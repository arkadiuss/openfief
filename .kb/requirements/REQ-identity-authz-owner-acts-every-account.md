---
title: Under either implementation, the owner must act on every account.
status: open
text_ref: design-decisions:D-163#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-163-4
  - lane:strict
semantic_text: Under either implementation, the owner must act on every account.
logic_claims:
  - CLAIM-3FE7E1BBC2CE6175
semantic_clauses:
  - Under either implementation, the owner must act on every account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3e0fc41301a7405badc0a86a1a623586bc87104b17e05145a370f5743220411c
semantic_inventory:
  - claim_key: CLAIM-3FE7E1BBC2CE6175
    claim_text: Under either implementation, the owner must act on every account
    role: normative
    status: modeled
    span:
      start: 0
      end: 64
firmness: firm
required_evidence: asserted
origin:
  kind: agent
  ref: The author approved splitting D-163's last sentence into one requirement per clause, with these wordings, in chat with the orchestrator, 2026-11-21
  approved_by: author, 2026-11-21
  recorded_at: '2026-11-21T22:19:15.123Z'
id: REQ-identity-authz-owner-acts-every-account
type: req
---
Under either implementation, the owner must act on every account.

## Context

Author decision D-163, clause 4 of its last sentence, as amended on main (5317aac): the owner acts on every account under either implementation of the authorization interface, the policy over the grants table or SpiceDB (D-164). The interim framing "until SpiceDB answers, its only implementation" is gone, because GrantPolicy, the default `--authz=policy`, now answers beside SpiceDB. The author approved this in-place reword in chat with the orchestrator on 2026-11-21; the requirement keeps its id, text_ref, history and links. D-65 has the owner hold every account and alone change anything. identityapp.Policy, over the pure identity.MayAct, answers for the owner under the interim policy, GrantPolicy defers to it, and the zed assertions in authz/validation.yaml show the same under SpiceDB. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> **[D-163]** Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account, and totals, the allocation and performance are computed over the accounts the caller may view. Under either implementation, the policy over the grants table or SpiceDB (D-164), the owner acts on every account; viewers, script tokens and view plugins view exactly the accounts their grants name (D-209); and demo sessions hold the household grant.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-163#4
