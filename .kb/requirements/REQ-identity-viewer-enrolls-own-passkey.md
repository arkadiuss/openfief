---
title: A viewer who opens the owner's link enrolls their own passkey, which creates the person, binds the passkey to them, makes the grants the link carried and signs them in with a read session, in one transaction with an audit entry for the link's use, the person, the passkey, each grant and the sign-in; the link is then spent.
status: open
firmness: default
required_evidence: unverified
semantic_text: A viewer who opens the owner's link enrolls their own passkey, which creates the person, binds the passkey to them, makes the grants the link carried and signs them in with a read session, in one transaction with an audit entry for the link's use, the person, the passkey, each grant and the sign-in; the link is then spent.
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b117c1909343c391030bd6b864c793f9db730ea733722817d67d85ad08230b07
semantic_inventory:
  - claim_key: CLAIM-DC43ADB85691B8AB
    claim_text: A viewer who opens the owner's link enrolls their own passkey, which creates the person, binds the passkey to them, makes the grants the link carried and signs them in with a read session, in one transaction with an audit entry for the link's use, the person, the passkey, each grant and the sign-in
    role: descriptive
    status: missing
    span:
      start: 0
      end: 299
  - claim_key: CLAIM-285F8CEB08C973F0
    claim_text: the link is then spent
    role: descriptive
    status: missing
    span:
      start: 301
      end: 323
logic_claims:
  - CLAIM-DC43ADB85691B8AB
  - CLAIM-285F8CEB08C973F0
id: REQ-identity-viewer-enrolls-own-passkey
type: req
---
A viewer who opens the owner's link enrolls their own passkey, which creates the person, binds the passkey to them, makes the grants the link carried and signs them in with a read session, in one transaction with an audit entry for the link's use, the person, the passkey, each grant and the sign-in; the link is then spent.

## Context

The core makes the viewer's ID and WebAuthn user handle when the ceremony begins, so their passkey carries their own handle and a later login names them. The grants reach SpiceDB after the commit, as any grant's do. Decisions: D-65, D-26, D-210, D-211.

## Source

> S6 task brief and the permissions plan's author and design-session answers (Q6, Q7, Q8), 2026-11-13.