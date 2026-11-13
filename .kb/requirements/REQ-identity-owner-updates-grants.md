---
title: The owner adds and removes the grants of a script token or a view plugin in one call, each change an audit entry, after checking the holder and every target to add, and lists the view plugins that read with their own identity with their grants.
status: open
origin:
  kind: agent
  ref: 'S5 task brief: the Access page, script tokens and view plugins (D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T18:00:00.000Z'
firmness: default
required_evidence: unverified
semantic_text: The owner adds and removes the grants of a script token or a view plugin in one call, each change an audit entry, after checking the holder and every target to add, and lists the view plugins that read with their own identity with their grants.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 66ce5c0a704afd6dc71e3692a4dd173dfee05357b778108df1ee09c9d0fa1ae4
semantic_inventory:
  - claim_key: CLAIM-DD2B7B97648E4DAC
    claim_text: The owner adds and removes the grants of a script token or a view plugin in one call, each change an audit entry, after checking the holder and every target to add, and lists the view plugins that read with their own identity with their grants
    role: descriptive
    status: missing
    span:
      start: 0
      end: 243
logic_claims:
  - CLAIM-DD2B7B97648E4DAC
id: REQ-identity-owner-updates-grants
type: req
---
The owner adds and removes the grants of a script token or a view plugin in one call, each change an audit entry, after checking the holder and every target to add, and lists the view plugins that read with their own identity with their grants.

## Context

UpdateGrants and ListViewPlugins are owner-session calls on the AuthService; a read (demo) session is refused. The plugins listed are those that hold or ask for the view capability, less the one that serves browsers and any being removed.