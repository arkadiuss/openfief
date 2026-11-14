---
title: Revoking a viewer and changing a viewer's grants need no fresh passkey check, issuing the invitation link does (D-183); each is audited.
status: open
origin:
  kind: agent
  ref: 'S6b task brief: the Access page''s People section and the viewers end-to-end test (D-65, D-183, D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T22:00:00.000Z'
id: REQ-ui-access-people-revoke-grant-no-passkey-check
type: req
firmness: default
semantic_text: Revoking a viewer and changing a viewer's grants need no fresh passkey check, issuing the invitation link does (D-183); each is audited.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f8815bd0b95faa477eba0acc51c7d93de46e01b9cb60707435d9c8fa2fb5cf34
semantic_inventory:
  - claim_key: CLAIM-E57EDE056FEC23A1
    claim_text: Revoking a viewer and changing a viewer's grants need no fresh passkey check, issuing the invitation link does (D-183)
    role: descriptive
    status: missing
    span:
      start: 0
      end: 118
  - claim_key: CLAIM-B62CF8ACA583CD9F
    claim_text: each is audited
    role: descriptive
    status: missing
    span:
      start: 120
      end: 135
logic_claims:
  - CLAIM-E57EDE056FEC23A1
  - CLAIM-B62CF8ACA583CD9F
---
Revoking a viewer and changing a viewer's grants need no fresh passkey check, issuing the invitation link does (D-183); each is audited.

## Context

The author decided on 2026-11-13 that D-183 covers enrolling or removing a passkey, issuing an enrollment link and creating a script token, and that revoking a viewer or changing a grant only takes access away or is the owner's plain choice and needs no passkey check. The core's calls already behave so; the page must not ask for one.

## Source

Author decision of 2026-11-13, recorded in the S6b task brief.