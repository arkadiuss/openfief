---
title: The UI plugin's Turnstile config is both or neither, its secret is read on each call and its manifest asks for one egress host from config
status: active
sourceFile: plugins/ui/turnstile_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-14T22:47:03.199Z'
id: TEST-plugins-ui-turnstile
type: test
---
Table tests of the plugin's config (site key and siteverify URL together, a placeholder or a short key refused, none leaves Turnstile off), the secret read from its file on every call with errors that name no secret, and the manifest: no fixed egress host, the host of turnstileVerifyUrl from config and the one optional Secret.