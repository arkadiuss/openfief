---
title: Under either authorization mode, the owner acts on every account; a viewer, a script token and a view plugin view exactly the accounts their grants name; and a demo session holds the household grant.
status: open
firmness: firm
required_evidence: asserted
text_ref: design-decisions:D-163#4-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-163-4-6
  - lane:strict
semantic_text: Under either authorization mode, the owner acts on every account; a viewer, a script token and a view plugin view exactly the accounts their grants name; and a demo session holds the household grant.
semantic_clauses:
  - Under either authorization mode, the owner acts on every account
  - a viewer, a script token and a view plugin view exactly the accounts their grants name
  - a demo session holds the household grant
logic_claims:
  - CLAIM-1DD318FE80B57149
  - CLAIM-DD70B02A66A2F3EE
  - CLAIM-7B2F2C65BF07D7CC
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_inventory:
  - claim_key: CLAIM-1DD318FE80B57149
    claim_text: Under either authorization mode, the owner acts on every account
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 64
  - claim_key: CLAIM-DD70B02A66A2F3EE
    claim_text: a viewer, a script token and a view plugin view exactly the accounts their grants name
    role: descriptive
    status: modeled
    span:
      start: 66
      end: 152
  - claim_key: CLAIM-7B2F2C65BF07D7CC
    claim_text: a demo session holds the household grant
    role: descriptive
    status: modeled
    span:
      start: 158
      end: 198
origin:
  kind: agent
  ref: The author approved superseding REQ-identity-authz-readers-view-every-account with this wording from the design session, in chat with the orchestrator, 2026-11-21; D-163 as amended in 5317aac
  approved_by: author, 2026-11-21
  recorded_at: '2026-11-21T21:47:30.831Z'
semantic_source_hash: 13e06f5ba86fc37b3ca31b431aa91a4d83043a0a915c602d331dd20115310409
id: REQ-identity-authz-grants-hold-under-either-mode
type: req
---
Under either authorization mode, the owner acts on every account; a viewer, a script token and a view plugin view exactly the accounts their grants name; and a demo session holds the household grant.

## Context

Author decision D-163 as amended on main (5317aac): its last sentence no longer says that until SpiceDB answers, every other holder of read access views every account. GrantPolicy, the default `--authz=policy`, limits a viewer, a script token and a view plugin to the accounts their grants name, as SpiceDB does (D-209), so D-163 now says what both implementations do. This requirement supersedes REQ-identity-authz-readers-view-every-account, which quotes the old interim sentence; the author approved the supersession in chat with the orchestrator on 2026-11-21, with this wording from the design session. The importer numbers the sentence's three semicolon-separated clauses 4 to 6, so the text_ref names that range; clause 4 is also the one REQ-identity-authz-owner-acts-every-account cites.

Evidence: unit and integration tests only, so the evidence floor is asserted. TestGrantPolicy (TEST-identity-grant-policy) shows a viewer, a script token and a view plugin viewing exactly their granted accounts and changing none, the owner acting on every account, and a demo session viewing every account; the zed assertions in authz/validation.yaml (TEST-identity-authz-schema) show the same under SpiceDB, including the household grant every demo session holds. TEST-identity-access and TEST-identity-app-access cover identity.MayAct and Policy, the answers for the owner and a demo session that GrantPolicy defers to.

## Why this changed

The old requirement held only while the policy let every token and plugin view every account. GrantPolicy now enforces grants for viewers, script tokens and view plugins too (bc73b2d, D-209), after existing tokens and plugins were migrated to explicit household grants, so the interim rule no longer holds under either authorization mode.

## Source

> **[D-163]** Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account, and totals, the allocation and performance are computed over the accounts the caller may view. Under either implementation, the policy over the grants table or SpiceDB (D-164), the owner acts on every account; viewers, script tokens and view plugins view exactly the accounts their grants name (D-209); and demo sessions hold the household grant.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-163#4-6
