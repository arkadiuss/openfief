---
title: Origins, sessions, scopes, passkeys, enrollment codes and ceremonies follow their pure rules
status: active
sourceFile: internal/identity/auth_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:34.672Z'
id: TEST-identity-auth-domain
type: test
---
Table tests of the pure identity rules with a fixed clock and synthetic example.com hosts: ParseOrigin accepts https hosts and http only on localhost and refuses IP literals, paths and uppercase; the allow-list refuses other origins; NewSession enforces that an authenticator grants at most its scope; Session.Check orders revoked, absolute expiry, idle expiry, wrong host and scope; Touched writes at most once a minute; session tokens and enrollment codes parse back and never echo the secret in errors; Passkey.Used refuses a signature counter that did not go up unless both are zero; enrollment codes and ceremonies expire and refuse another origin or kind.