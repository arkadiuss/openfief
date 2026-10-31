---
title: The egress proxy's address check refuses private and internal ranges for IPv4 and IPv6
status: active
sourceFile: internal/plugins/egress/address_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:31:54.707Z'
id: TEST-plugins-egress-address
type: test
verification_scope: unit
---
Table tests call CheckAddress on parsed addresses: public resolver addresses, addresses just outside 172.16/12 and the carrier-grade NAT range, and NAT64, 6to4 and IPv4-mapped forms of a public address are allowed; unspecified, this-network, RFC 1918, loopback, link-local, the metadata address 169.254.169.254, carrier-grade NAT, protocol-assignment, documentation, benchmarking, multicast, reserved and broadcast IPv4 addresses, and unspecified, loopback, unique-local (including fd00:ec2::254), link-local with and without zone, site-local, multicast, documentation, Teredo, discard and local-use NAT64 IPv6 addresses are refused, as are private IPv4 addresses hidden in IPv4-mapped, IPv4-compatible, NAT64 and 6to4 form, and the zero address. A false pass would be a mapped or embedded private address slipping through; each form has a case.