---
title: Under either implementation, viewers, script tokens and view plugins must view exactly the accounts their grants name.
status: open
firmness: firm
required_evidence: asserted
text_ref: design-decisions:D-163#5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-163-5
  - lane:strict
semantic_text: Under either implementation, viewers, script tokens and view plugins must view exactly the accounts their grants name.
semantic_clauses:
  - Under either implementation, viewers, script tokens and view plugins must view exactly the accounts their grants name.
logic_claims:
  - CLAIM-52EFC3630AFC063F
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_inventory:
  - claim_key: CLAIM-52EFC3630AFC063F
    claim_text: Under either implementation, viewers, script tokens and view plugins must view exactly the accounts their grants name
    role: normative
    status: modeled
    span:
      start: 0
      end: 117
origin:
  kind: agent
  ref: The author approved splitting D-163's last sentence into one requirement per clause, with these wordings, in chat with the orchestrator, 2026-11-21
  approved_by: author, 2026-11-21
  recorded_at: '2026-11-21T22:19:20.221Z'
semantic_source_hash: 59ebe8bd12beeb7e0207d5b1d188a76502e7f788bdaa8df19ec37273b7c87a38
id: REQ-identity-authz-grants-hold-under-either-mode
type: req
approval_ref: The author approved this Context-only update for SpiceDB as the default (D-164) in chat with the orchestrator, 2026-11-22
---
Under either implementation, viewers, script tokens and view plugins must view exactly the accounts their grants name.

## Context

Author decision D-163, clause 5 of its last sentence, as amended on main (5317aac). GrantPolicy, the opt-out `--authz=policy`, limits a viewer, a script token and a view plugin to the accounts their grants name, as SpiceDB, the default `--authz=spicedb` (D-164), does (D-209), so D-163 says what both implementations do. This requirement supersedes REQ-identity-authz-readers-view-every-account, which quotes the old interim sentence; the author approved the supersession in chat with the orchestrator on 2026-11-21. It first covered the whole sentence under the range text_ref #4-6; the author then approved one requirement per clause (the owner's is REQ-identity-authz-owner-acts-every-account, the demo's REQ-identity-authz-demo-holds-household), because one requirement would hide the thin demo clause. Evidence: unit and integration tests only, so the evidence floor is asserted. TestGrantPolicy (TEST-identity-grant-policy) shows a viewer, a script token and a view plugin viewing exactly their granted accounts and changing none, and the zed assertions in authz/validation.yaml (TEST-identity-authz-schema) show the same under SpiceDB.

## Why this changed

The old requirement held only while the policy let every token and plugin view every account. GrantPolicy now enforces grants for viewers, script tokens and view plugins too (bc73b2d, D-209), after existing tokens and plugins were migrated to explicit household grants, so the interim rule no longer holds under either implementation.

## Source

> **[D-163]** Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account, and totals, the allocation and performance are computed over the accounts the caller may view. Under either implementation, the policy over the grants table or SpiceDB (D-164), the owner acts on every account; viewers, script tokens and view plugins view exactly the accounts their grants name (D-209); and demo sessions hold the household grant.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-163#5
