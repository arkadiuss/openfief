---
title: MayMountSecrets lets a mounting or older instance keep a Secret and refuses any other
status: active
sourceFile: internal/plugins/secret_holder_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T08:22:02.182Z'
id: TEST-plugins-secret-holder
type: test
---
TestMayMountSecrets is a table over plugins.MayMountSecrets with synthetic holders: a Secret nobody mounts may be mounted, one another instance mounts is refused with ErrSecretHeld, the holder itself and the older of two that both mount it keep it, the younger is refused, and unrelated Secrets pass.