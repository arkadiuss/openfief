---
title: 'The operator must record each plugin''s signature trust in its status in full: verified, signed by OpenFief''s CI, by a catalog publisher, by the owner, by a trusted signer of unknown kind, refused, or not checked.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-87-1
  - lane:strict
text_ref: design-decisions:D-87#1
origin:
  kind: import
  ref: design-decisions:D-87#1
  recorded_at: '2026-10-29T08:27:27.606Z'
semantic_text: 'The operator must record each plugin''s signature trust in its status in full: verified, signed by OpenFief''s CI, by a catalog publisher, by the owner, by a trusted signer of unknown kind, refused, or not checked.'
logic_claims:
  - CLAIM-7DAAFBDE0635CA7C
semantic_clauses:
  - 'The operator must record each plugin''s signature trust in its status in full: verified, signed by OpenFief''s CI, by a catalog publisher, by the owner, by a trusted signer of unknown kind, refused, or not checked.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 5806f6e640579cb54139a51e8e6b3c8a76f27596404e82ca7f487c42e74debfd
semantic_inventory:
  - claim_key: CLAIM-7DAAFBDE0635CA7C
    claim_text: 'The operator must record each plugin''s signature trust in its status in full: verified, signed by OpenFief''s CI, by a catalog publisher, by the owner, by a trusted signer of unknown kind, refused, or not checked'
    role: normative
    status: modeled
    span:
      start: 0
      end: 211
id: REQ-operator-plugin-trust-status
type: req
---
The operator must record each plugin's signature trust in its status in full: verified, signed by OpenFief's CI, by a catalog publisher, by the owner, by a trusted signer of unknown kind, refused, or not checked.

## Context

D-56 has the operator record in each Plugin's status whether its image is verified or only signed by a trusted signer (REQ-plugins-verified-status), and D-45 names the three kinds of trusted signer. D-87 has the operator record the trust in full, so the status distinguishes each kind of signer, an unknown kind, a refusal and an image not yet checked. REQ-ui-plugin-trust-badges and REQ-ui-plugin-trust-badge-mapping record how the UI groups these levels. The author recorded D-86 and D-87 after the design session that answered the implementation's questions.

## Source

> The operator records each plugin's signature trust in its status in full: verified, signed by OpenFief's CI, by a catalog publisher, by the owner, by a trusted signer of unknown kind, refused, or not checked. The UI groups them into four badges, with the exact signer in the details: *Verified* (in the catalog and reviewed, D-44), *Yours* (one of the owner's signers, D-45), *Unreviewed* (another trusted signer, not in the catalog) and *Blocked* (refused or not checked yet).

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-87#1
