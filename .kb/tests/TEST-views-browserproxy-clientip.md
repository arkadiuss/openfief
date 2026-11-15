---
title: The front door finds the address only from trusted proxies and hands the plugin one value
status: active
sourceFile: internal/views/browserproxy/clientip_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-14T20:37:54.774Z'
id: TEST-views-browserproxy-clientip
type: test
---
Parsing and resolving: X-Forwarded-For read from its right end, single-address headers, a trusted proxy that says nothing, an untrusted peer's header never believed, malformed entries; and through the proxy, the plugin sees the address found and never a value the browser sent, with the header left out when unknown.