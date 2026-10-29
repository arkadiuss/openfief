---
title: 'The UI must group plugins'' signature trust into exactly four badges: Verified, Yours, Unreviewed and Blocked.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-87-2
  - lane:strict
text_ref: design-decisions:D-87#2
origin:
  kind: import
  ref: design-decisions:D-87#2
  recorded_at: '2026-10-29T08:28:05.210Z'
semantic_text: 'The UI must group plugins'' signature trust into exactly four badges: Verified, Yours, Unreviewed and Blocked.'
logic_claims:
  - CLAIM-E07F711CA8AFC8A9
semantic_clauses:
  - 'The UI must group plugins'' signature trust into exactly four badges: Verified, Yours, Unreviewed and Blocked.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 4ec7bbffeff33a6bdb73897b33ef72480b9dde21bab0d323e576fe9dd0f24cbc
semantic_inventory:
  - claim_key: CLAIM-E07F711CA8AFC8A9
    claim_text: 'The UI must group plugins'' signature trust into exactly four badges: Verified, Yours, Unreviewed and Blocked'
    role: normative
    status: modeled
    span:
      start: 0
      end: 108
id: REQ-ui-plugin-trust-badges
type: req
---
The UI must group plugins' signature trust into exactly four badges: Verified, Yours, Unreviewed and Blocked.

## Context

D-56 has the UI show each plugin's signature status, and D-87 has the operator record that trust in full, in seven levels (REQ-operator-plugin-trust-status). The author groups those levels into four badges for the owner and visitors; REQ-ui-plugin-trust-badge-mapping records which level goes to which badge and REQ-ui-plugin-trust-badge-signer keeps the exact signer in the details. REQ-ui-installed-plugins, which the author has not reviewed, shows a trust level per plugin without naming these badges. The author recorded D-86 and D-87 after the design session that answered the implementation's questions.

## Source

> The operator records each plugin's signature trust in its status in full: verified, signed by OpenFief's CI, by a catalog publisher, by the owner, by a trusted signer of unknown kind, refused, or not checked. The UI groups them into four badges, with the exact signer in the details: *Verified* (in the catalog and reviewed, D-44), *Yours* (one of the owner's signers, D-45), *Unreviewed* (another trusted signer, not in the catalog) and *Blocked* (refused or not checked yet).

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-87#2
