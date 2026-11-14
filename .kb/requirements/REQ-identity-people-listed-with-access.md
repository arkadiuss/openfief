---
title: The owner lists the people, the owner first and each viewer with the grants in force, the passkeys that work and the sessions still valid, and updates a viewer's grants as a token's.
status: open
firmness: default
required_evidence: unverified
semantic_text: The owner lists the people, the owner first and each viewer with the grants in force, the passkeys that work and the sessions still valid, and updates a viewer's grants as a token's.
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 58da03f128d812ea8aa25692f7862c76edb403355a6e38f07b12418c37bedafc
semantic_inventory:
  - claim_key: CLAIM-3F74BD936CF35670
    claim_text: The owner lists the people, the owner first and each viewer with the grants in force, the passkeys that work and the sessions still valid, and updates a viewer's grants as a token's
    role: descriptive
    status: missing
    span:
      start: 0
      end: 181
logic_claims:
  - CLAIM-3F74BD936CF35670
id: REQ-identity-people-listed-with-access
type: req
---
The owner lists the people, the owner first and each viewer with the grants in force, the passkeys that work and the sessions still valid, and updates a viewer's grants as a token's.

## Context

ListPeople and UpdateGrants of a person holder are owner-session calls; a revoked viewer cannot be granted. Decisions: D-210.

## Source

> S6 task brief and the permissions plan's author and design-session answers (Q6, Q7, Q8), 2026-11-13.