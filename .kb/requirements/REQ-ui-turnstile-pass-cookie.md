---
title: The Turnstile pass is a cookie tied to the visitor's session and signed with a key derived from the install's Turnstile secret, HttpOnly, Secure, SameSite=Strict and scoped to the Ask box's path, holding only its version, the check time, its end and the signature, and never logged; a cookie that is malformed, forged, made for another session or under another key falls back to the on-demand check and never shows an error.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Turnstile pass is a cookie tied to the visitor's session and signed with a key derived from the install's Turnstile secret, HttpOnly, Secure, SameSite=Strict and scoped to the Ask box's path, holding only its version, the check time, its end and the signature, and never logged; a cookie that is malformed, forged, made for another session or under another key falls back to the on-demand check and never shows an error.
origin:
  kind: agent
  ref: 'The design session''s amendment of D-208 (docs/DESIGN.md): a visitor who passes Turnstile isn''t checked again for 30 minutes after their last model question, and at most 2 hours from the check, through a signed cookie tied to their session that holds nothing else; the rate limits and the budget still count every question (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
tags:
  - turnstile
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 895430f5188b6f55bd608119a3ae2b287c5a1d360916febf219223f80862e205
logic_claims:
  - CLAIM-AA6F26A16D2A228B
  - CLAIM-BAF51DEA7AB23F6F
semantic_inventory:
  - claim_key: CLAIM-AA6F26A16D2A228B
    claim_text: The Turnstile pass is a cookie tied to the visitor's session and signed with a key derived from the install's Turnstile secret, HttpOnly, Secure, SameSite=Strict and scoped to the Ask box's path, holding only its version, the check time, its end and the signature, and never logged
    role: descriptive
    status: ontology_gap
    span:
      start: 0
      end: 281
    subjects:
      - turnstile.pass_cookie
    reason: It states how the built-in UI behaves; no predicate of the ontology states it yet.
  - claim_key: CLAIM-BAF51DEA7AB23F6F
    claim_text: a cookie that is malformed, forged, made for another session or under another key falls back to the on-demand check and never shows an error
    role: descriptive
    status: ontology_gap
    span:
      start: 283
      end: 423
    subjects:
      - turnstile.pass_cookie
    reason: It states how the built-in UI behaves; no predicate of the ontology states it yet.
rationale: 'D-208, as amended, says the pass is a signed cookie tied to the visitor''s session that holds nothing else. The design brief adds the flags, the key (derived, not logged), the content (no user identifier other than the binding) and the fallback: a cookie that fails to verify behaves as none and never errors visibly.'
id: REQ-ui-turnstile-pass-cookie
type: req
---
The Turnstile pass is a cookie tied to the visitor's session and signed with a key derived from the install's Turnstile secret, HttpOnly, Secure, SameSite=Strict and scoped to the Ask box's path, holding only its version, the check time, its end and the signature, and never logged; a cookie that is malformed, forged, made for another session or under another key falls back to the on-demand check and never shows an error.

## Context

D-208, as amended, says the pass is a signed cookie tied to the visitor's session that holds nothing else. The design brief adds the flags, the key (derived, not logged), the content (no user identifier other than the binding) and the fallback: a cookie that fails to verify behaves as none and never errors visibly.

## Source

> D-208 (docs/DESIGN.md), as amended: A visitor who passes Turnstile isn't checked again for 30 minutes after their last model question, and at most 2 hours from the check, through a signed cookie tied to their session that holds nothing else; the rate limits and the budget still count every question.
