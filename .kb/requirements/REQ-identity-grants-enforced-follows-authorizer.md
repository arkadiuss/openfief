---
title: The core must report grants as enforced, in its lists of script tokens, view plugins and people, exactly when its authorizer limits holders to their grants, which SpiceDB and the policy over the grants table (--authz=policy) both do.
status: open
firmness: default
semantic_text: The core must report grants as enforced, in its lists of script tokens, view plugins and people, exactly when its authorizer limits holders to their grants, which SpiceDB and the policy over the grants table (--authz=policy) both do.
origin:
  kind: agent
  ref: 'Grants.Enforced (low), internal/identity/app/grants.go: Enforced() returns g.proj != nil, and its comment says the policy lets every holder view every account. Both have been stale since GrantPolicy started enforcing grants. Use identityapp.EnforcesTokenGrants(authorizer) or the equivalent, fix the comment, and test both modes. (the owner''s items for alpha.2, from the reviews of main 0bfeb62)'
  recorded_at: '2026-11-27T08:26:28.982Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fbfeed829f6d0d3a5e29a91d918c8afbd922ad700bb42ddde1c7ef2d9eae2961
semantic_inventory:
  - claim_key: CLAIM-3C2E59C80A781149
    claim_text: The core must report grants as enforced, in its lists of script tokens, view plugins and people, exactly when its authorizer limits holders to their grants, which SpiceDB and the policy over the grants table (--authz=policy) both do
    role: normative
    status: missing
    span:
      start: 0
      end: 232
    subjects:
      - identity.grants_enforced_report
logic_claims:
  - CLAIM-3C2E59C80A781149
id: REQ-identity-grants-enforced-follows-authorizer
type: req
---
The core must report grants as enforced, in its lists of script tokens, view plugins and people, exactly when its authorizer limits holders to their grants, which SpiceDB and the policy over the grants table (--authz=policy) both do.

## Context

Grants.Enforced answered whether a SpiceDB syncer was wired, so under --authz=policy the Access page wrongly warned that every token reads every account, although GrantPolicy has enforced grants since D-209 (finding of the reviews of main 0bfeb62). It now asks the authorizer in use (identityapp.EnforcesTokenGrants).

## Source

> Grants.Enforced (low), internal/identity/app/grants.go: Enforced() returns g.proj != nil, and its comment says the policy lets every holder view every account. Both have been stale since GrantPolicy started enforcing grants. Use identityapp.EnforcesTokenGrants(authorizer) or the equivalent, fix the comment, and test both modes. (the owner's items for alpha.2, from the reviews of main 0bfeb62)