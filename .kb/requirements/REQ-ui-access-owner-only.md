---
title: 'The Access page and its forms are the owner''s: a read session never sees the page or its link, and is refused its forms, which never reach the core.'
status: open
origin:
  kind: agent
  ref: 'S5 task brief: the Access page, script tokens and view plugins (D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T18:00:00.000Z'
firmness: default
required_evidence: unverified
semantic_text: 'The Access page and its forms are the owner''s: a read session never sees the page or its link, and is refused its forms, which never reach the core.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: da1b03cddc3742e5eb7ebf1c66169c89ccd138646b668b227cd241683441779b
semantic_inventory:
  - claim_key: CLAIM-A16113643101244B
    claim_text: 'The Access page and its forms are the owner''s: a read session never sees the page or its link, and is refused its forms, which never reach the core'
    role: descriptive
    status: missing
    span:
      start: 0
      end: 147
logic_claims:
  - CLAIM-A16113643101244B
id: REQ-ui-access-owner-only
type: req
---
The Access page and its forms are the owner's: a read session never sees the page or its link, and is refused its forms, which never reach the core.

## Context

D-210 puts the Access page in Settings and D-209 makes grants the owner's alone. Both authorization modes enforce grants: SpiceDB, the default (D-164), and the policy, the owner's --authz=policy opt-out, which GrantPolicy answers from the grants table (D-209, D-163). So a grant the page stores limits its holder at once under either mode, and the page carries no note that grants have no effect: it had one for a core that answered with no grants enforced, which no longer exists, and it wrongly showed it under --authz=policy until Grants.Enforced followed the authorizer.