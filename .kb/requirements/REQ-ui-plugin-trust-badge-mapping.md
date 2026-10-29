---
title: The UI must badge a plugin Verified when it is in the catalog and reviewed, Yours when one of the owner's signers signed it, Unreviewed when another trusted signer signed it and it is not in the catalog, and Blocked when it is refused or not checked yet.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-87-4
  - lane:strict
text_ref: design-decisions:D-87#4
origin:
  kind: import
  ref: design-decisions:D-87#4
  recorded_at: '2026-10-29T08:29:33.144Z'
semantic_text: The UI must badge a plugin Verified when it is in the catalog and reviewed, Yours when one of the owner's signers signed it, Unreviewed when another trusted signer signed it and it is not in the catalog, and Blocked when it is refused or not checked yet.
logic_claims:
  - CLAIM-95B4B70807F062FF
semantic_clauses:
  - The UI must badge a plugin Verified when it is in the catalog and reviewed, Yours when one of the owner's signers signed it, Unreviewed when another trusted signer signed it and it is not in the catalog, and Blocked when it is refused or not checked yet.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 073d135492305f9afff7b2af9fa33254a1e98792ea9818a9842041f02161539e
semantic_inventory:
  - claim_key: CLAIM-95B4B70807F062FF
    claim_text: The UI must badge a plugin Verified when it is in the catalog and reviewed, Yours when one of the owner's signers signed it, Unreviewed when another trusted signer signed it and it is not in the catalog, and Blocked when it is refused or not checked yet
    role: normative
    status: modeled
    span:
      start: 0
      end: 253
id: REQ-ui-plugin-trust-badge-mapping
type: req
---
The UI must badge a plugin Verified when it is in the catalog and reviewed, Yours when one of the owner's signers signed it, Unreviewed when another trusted signer signed it and it is not in the catalog, and Blocked when it is refused or not checked yet.

## Context

D-87 groups the seven trust levels the operator records (REQ-operator-plugin-trust-status) into the four badges of REQ-ui-plugin-trust-badges. Verified is D-44's reviewed catalog entry and Yours is D-45's owner signer. Unreviewed covers every other trusted signer whose image is not in the catalog; read with D-45's signer kinds, that is an image signed by OpenFief's CI, by a catalog publisher or by a trusted signer of unknown kind without a catalog entry pinning its digest, which is how the property fact records it. Blocked covers a refused image and one not checked yet. The author recorded D-86 and D-87 after the design session that answered the implementation's questions.

## Source

> The operator records each plugin's signature trust in its status in full: verified, signed by OpenFief's CI, by a catalog publisher, by the owner, by a trusted signer of unknown kind, refused, or not checked. The UI groups them into four badges, with the exact signer in the details: *Verified* (in the catalog and reviewed, D-44), *Yours* (one of the owner's signers, D-45), *Unreviewed* (another trusted signer, not in the catalog) and *Blocked* (refused or not checked yet).

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-87#4
