---
title: 'An enrollment code is spent only before its expiry: the statement that uses it up also checks that it has not expired, so a registration begun with the code before it expired and finished after is refused, and the code stays unused.'
status: open
firmness: default
semantic_text: 'An enrollment code is spent only before its expiry: the statement that uses it up also checks that it has not expired, so a registration begun with the code before it expired and finished after is refused, and the code stays unused.'
origin:
  kind: agent
  ref: 'Security review before the alpha.2 cut, known limit #4: re-check an invite''s expiry atomically when registration completes (the coordinator''s follow-up, 2026-11-21)'
  recorded_at: '2026-11-21T18:38:12.129Z'
rationale: A link whose expiry the owner chose must not be usable after it, even by a ceremony begun just before; checking it only at the start left a five-minute window.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 894cdce41a335ef26d6ecd8cd911112a032ee64ec402ef27f21f80fe0a58c607
semantic_inventory:
  - claim_key: CLAIM-59D66BB0EFFABAF2
    claim_text: 'An enrollment code is spent only before its expiry: the statement that uses it up also checks that it has not expired, so a registration begun with the code before it expired and finished after is refused, and the code stays unused'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 231
    subjects:
      - identity.enrollment_code_expiry_at_use
    reason: It states when a stored code may be spent; no predicate of the ontology states a time condition on a state change yet.
logic_claims:
  - CLAIM-59D66BB0EFFABAF2
id: REQ-identity-enrollment-code-expiry-at-use
type: req
---
An enrollment code is spent only before its expiry: the statement that uses it up also checks that it has not expired, so a registration begun with the code before it expired and finished after is refused, and the code stays unused.

## Context

The code's expiry was checked when the registration began, and only used and cancelled were checked when it finished, so a ceremony begun a second before the expiry could finish up to its five minutes after it. Checking the expiry in the same UPDATE that spends the code, in the registration's transaction, closes that window without a race.

## Source

> Security review of the new surfaces before the alpha.2 cut: the invites (the design session's exit criterion (c), 2026-11-21); the coordinator asked to fix known limit #4.