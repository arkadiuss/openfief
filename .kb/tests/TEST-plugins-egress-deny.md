---
title: The egress proxy's extra deny ranges parse and are refused
status: active
sourceFile: internal/plugins/egress/deny_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T00:10:28.181Z'
id: TEST-plugins-egress-deny
type: test
---
TestParseDenyList reads synthetic deny ranges and node addresses, masked, once each, refusing what doesn't parse, zoned node addresses included; TestCheckAddressDenying refuses a synthetic node address directly, IPv4-mapped and inside NAT64 and 6to4, and a global IPv6 cluster range, while another public address passes and a private one stays refused.