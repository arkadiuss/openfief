---
title: A grant names a holder and a target by IDs of fixed shapes, and an audit entry names exactly its holder then its target
status: active
sourceFile: internal/identity/grant_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T10:00:07.915Z'
id: TEST-identity-grant-domain
type: test
---
Table tests of identity.NewGrant, Grant.Revoked, GrantMatch and the grant audit entries, pure functions. A script token, a view plugin or a person is granted an account, a sleeve or the household; a holder named by a secret, a sentence, capitals or a name too long, a target named by a secret or with a space, and a household that is not home are refused with ErrInvalidGrant, whose message never quotes the holder's ID. An ending never precedes the grant's own time. A grant_created or grant_revoked entry names its holder, then its target, by kind and ID, and nothing else: no references, one, three, the target first, two holders and a session as the target are all refused. A false pass would be a grant that accepted a free-text ID, which the secret and sentence cases rule out.