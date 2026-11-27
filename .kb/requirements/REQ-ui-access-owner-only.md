---
title: 'The Access page and its forms are the owner''s: a read session never sees the page or its link, and is refused its forms, which never reach the core; while grants have no effect it says so.'
status: open
origin:
  kind: agent
  ref: 'S5 task brief: the Access page, script tokens and view plugins (D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T18:00:00.000Z'
firmness: default
required_evidence: unverified
semantic_text: 'The Access page and its forms are the owner''s: a read session never sees the page or its link, and is refused its forms, which never reach the core; while grants have no effect it says so.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 20a55e2b580a9a2d13bd3d59659aa8f6728830ed4c13ddada470f631042c0150
semantic_inventory:
  - claim_key: CLAIM-A16113643101244B
    claim_text: 'The Access page and its forms are the owner''s: a read session never sees the page or its link, and is refused its forms, which never reach the core'
    role: descriptive
    status: missing
    span:
      start: 0
      end: 147
  - claim_key: CLAIM-E097275747DE5F18
    claim_text: while grants have no effect it says so
    role: descriptive
    status: missing
    span:
      start: 149
      end: 187
logic_claims:
  - CLAIM-A16113643101244B
  - CLAIM-E097275747DE5F18
id: REQ-ui-access-owner-only
type: req
---
The Access page and its forms are the owner's: a read session never sees the page or its link, and is refused its forms, which never reach the core; while grants have no effect it says so.

## Context

D-210 puts the Access page in Settings and D-209 makes grants the owner's alone. Both authorization modes enforce grants: SpiceDB, the default (D-164), and the policy, the owner's --authz=policy opt-out, which GrantPolicy answers from the grants table (D-209, D-163). So a grant the page stores limits its holder at once under either mode, and the page's note that grants have no effect is shown only by a core whose authorizer reports that it does not enforce them, which neither does.