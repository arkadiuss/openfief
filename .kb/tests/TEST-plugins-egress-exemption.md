---
title: The refused-range exemption parses, covers exactly one host at exactly its addresses, and leaves everything else refused
status: active
sourceFile: internal/plugins/egress/exemption_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T23:20:13.174Z'
id: TEST-plugins-egress-exemption
type: test
---
TestParseExemption reads synthetic host=address lists (sorted, once each, IPv4-mapped addresses unmapped, the host in any case) and refuses no addresses, ranges, wildcards, IP addresses and ports as hosts, zoned addresses and a second host. TestCheckResolved, a table of synthetic hosts and addresses, passes the exempt host at its address, and refuses another host at that address, another address in the same range, the exempt host at another address or at a further one, a parent or subdomain of it, the exempt address as an IP literal, a host with no address, and every address with no exemption, with the node address overridden for the exempt host only.