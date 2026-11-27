---
title: A session's principal must be built in one place from the session, as the viewer whose session it is, as the demo for a session from the demo authenticator, and as the owner only for the owner's own session, and any other principal must be nobody, which the authorizer refuses, so a viewer's session is never the owner.
status: open
firmness: default
semantic_text: A session's principal must be built in one place from the session, as the viewer whose session it is, as the demo for a session from the demo authenticator, and as the owner only for the owner's own session, and any other principal must be nobody, which the authorizer refuses, so a viewer's session is never the owner.
origin:
  kind: agent
  ref: 'S1 (security, medium): a viewer resolves as the owner. Add one constructor in identity/app that builds a Principal from a Session, and use it everywhere. Make Caller() fail closed: HolderOwner only from an explicit owner scope (the session is not a viewer''s and not a demo), never as the fall-through. An unknown case returns an empty Caller that the authorizer refuses. (the owner''s items for alpha.2, from the reviews of main 0bfeb62)'
  recorded_at: '2026-11-27T09:04:51.519Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 55528cb334eb2793a8734b3ee887fec0d12170379488ec4ce10c8b2097b78cdb
semantic_inventory:
  - claim_key: CLAIM-2FD4AEEE4E1BD7B8
    claim_text: A session's principal must be built in one place from the session, as the viewer whose session it is, as the demo for a session from the demo authenticator, and as the owner only for the owner's own session, and any other principal must be nobody, which the authorizer refuses, so a viewer's session is never the owner
    role: normative
    status: missing
    span:
      start: 0
      end: 318
    subjects:
      - identity.session_principal
logic_claims:
  - CLAIM-2FD4AEEE4E1BD7B8
id: REQ-identity-principal-of-session
type: req
---
A session's principal must be built in one place from the session, as the viewer whose session it is, as the demo for a session from the demo authenticator, and as the owner only for the owner's own session, and any other principal must be nobody, which the authorizer refuses, so a viewer's session is never the owner.

## Context

The maintainer built its Principal by hand and left out the viewer, so Principal.Caller() fell through to the owner and a limited viewer was shown the household-wide note (finding S1, medium, of the reviews of main 0bfeb62). A hand-built principal that names no kind of session is now nobody.

## Source

> S1 (security, medium): a viewer resolves as the owner. Add one constructor in identity/app that builds a Principal from a Session, and use it everywhere. Make Caller() fail closed: HolderOwner only from an explicit owner scope (the session is not a viewer's and not a demo), never as the fall-through. An unknown case returns an empty Caller that the authorizer refuses. (the owner's items for alpha.2, from the reviews of main 0bfeb62)