---
title: Both core ports refuse compressed bodies, cap bodies by path, and bound request times
status: active
sourceFile: cmd/core/limits_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-02T19:45:10.299Z'
id: TEST-core-request-limits
type: test
---
TestPortsBoundBodies sends lazily made bodies through the browser and plugin port handlers to an API that drains them: the read API and an owner service stop just past 2 MiB with 413, the form proxied to the view plugin and the plugin port's statement upload stream pass 3 MiB, a declared length over the cap is refused unread, and compressed bodies get 415 on both ports. TestListenTimeouts starts both ports on loopback listeners and checks each server's read header, read and idle timeouts.