---
title: The Plugins page lists the installed plugins with their image digests and signers to a viewer, one granted nothing included, exactly as it does to anyone.
status: open
origin:
  kind: agent
  ref: 'author answers of 2026-11-13 recorded in the S7 task brief (S6c): invitations, Inbox and Plugins pages of a viewer'
  recorded_at: '2026-11-13T23:30:00.000Z'
firmness: default
semantic_text: The Plugins page lists the installed plugins with their image digests and signers to a viewer, one granted nothing included, exactly as it does to anyone.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0b42049c65e9c6ebff94d6c571429f8919418900d301d517a49cdbdfa9307374
semantic_inventory:
  - claim_key: CLAIM-0B8E74B1B9DE4242
    claim_text: The Plugins page lists the installed plugins with their image digests and signers to a viewer, one granted nothing included, exactly as it does to anyone
    role: descriptive
    status: missing
    span:
      start: 0
      end: 153
logic_claims:
  - CLAIM-0B8E74B1B9DE4242
id: REQ-ui-plugins-viewer-lists-installed
type: req
---
The Plugins page lists the installed plugins with their image digests and signers to a viewer, one granted nothing included, exactly as it does to anyone.

## Context

D-86 makes the list of installed plugins readable by any read API caller. The author reported on 2026-11-13 that a viewer must see it, any grants or none; a viewer granted nothing lands on the empty state of the Sleeves page, which left no way to the Plugins page, so that page now keeps the settings navigation.

## Source

D-86: The list of installed plugins, with their image digests and signers, is readable by any read API caller (D-83), including the demo's read-only visitors. Author answers of 2026-11-13.