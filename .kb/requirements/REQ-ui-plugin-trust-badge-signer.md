---
title: The UI must show a plugin's exact signer in the details of its trust badge.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-87-3
  - lane:strict
text_ref: design-decisions:D-87#3
origin:
  kind: import
  ref: design-decisions:D-87#3
  recorded_at: '2026-10-29T08:28:41.891Z'
semantic_text: The UI must show a plugin's exact signer in the details of its trust badge.
logic_claims:
  - CLAIM-3559AB1EFB8A3E50
semantic_clauses:
  - The UI must show a plugin's exact signer in the details of its trust badge.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 36ebbec3729654d52a4ef4815f8b65135ef8565814141456038ef78f6526f497
semantic_inventory:
  - claim_key: CLAIM-3559AB1EFB8A3E50
    claim_text: The UI must show a plugin's exact signer in the details of its trust badge
    role: normative
    status: modeled
    span:
      start: 0
      end: 74
id: REQ-ui-plugin-trust-badge-signer
type: req
---
The UI must show a plugin's exact signer in the details of its trust badge.

## Context

D-87 groups the seven trust levels the operator records (REQ-operator-plugin-trust-status) into four badges (REQ-ui-plugin-trust-badges), which lose which signer earned a plugin its badge. The author keeps that information one step away: the badge's details name the exact signer. D-86 makes the signers readable by any read API caller (REQ-views-installed-plugins-any-caller). The author recorded D-86 and D-87 after the design session that answered the implementation's questions.

## Source

> The operator records each plugin's signature trust in its status in full: verified, signed by OpenFief's CI, by a catalog publisher, by the owner, by a trusted signer of unknown kind, refused, or not checked. The UI groups them into four badges, with the exact signer in the details: *Verified* (in the catalog and reviewed, D-44), *Yours* (one of the owner's signers, D-45), *Unreviewed* (another trusted signer, not in the catalog) and *Blocked* (refused or not checked yet).

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-87#3
