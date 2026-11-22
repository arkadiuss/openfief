---
title: Under either implementation, demo sessions must hold the household grant.
status: open
firmness: firm
required_evidence: asserted
text_ref: design-decisions:D-163#6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-163-6
  - lane:strict
semantic_text: Under either implementation, demo sessions must hold the household grant.
semantic_clauses:
  - Under either implementation, demo sessions must hold the household grant.
logic_claims:
  - CLAIM-E60A70175EAA2ADF
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d767d230d656a9c4a81891566724181a114b205e7e8027509e4793d9d724c127
semantic_inventory:
  - claim_key: CLAIM-E60A70175EAA2ADF
    claim_text: Under either implementation, demo sessions must hold the household grant
    role: normative
    status: modeled
    span:
      start: 0
      end: 72
origin:
  kind: agent
  ref: The author approved splitting D-163's last sentence into one requirement per clause, with these wordings, in chat with the orchestrator, 2026-11-21
  approved_by: author, 2026-11-21
  recorded_at: '2026-11-21T22:19:24.313Z'
id: REQ-identity-authz-demo-holds-household
type: req
---
Under either implementation, demo sessions must hold the household grant.

## Context

Author decision D-163, clause 6 of its last sentence, as amended on main (5317aac), split out of the first superseding requirement by the author's approval in chat with the orchestrator on 2026-11-21. A demo session is the read-scoped session of the demo authenticator (D-68, D-137); it holds the household grant, so it views every account of the household, present and future (D-209), and changes none. Under the policy, GrantPolicy answers it as identity.MayAct does; under SpiceDB, `household:home#reader@demo_session:*` gives every demo session the grant, as the zed assertions in authz/validation.yaml show. TestGrantPolicy checks a demo session views every account, including one added after the session started, and changes none. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> **[D-163]** Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account, and totals, the allocation and performance are computed over the accounts the caller may view. Under either implementation, the policy over the grants table or SpiceDB (D-164), the owner acts on every account; viewers, script tokens and view plugins view exactly the accounts their grants name (D-209); and demo sessions hold the household grant.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-163#6
