---
title: 'The client events'' limit: an IPv6 address by its /64, an IPv4 address as itself, and the cap on every address together'
status: active
sourceFile: internal/views/ui/clientevents_test.go
verification_scope: unit
origin:
  kind: agent
  ref: Security review of the new surfaces before the alpha.2 cut, /client-events (2026-11-21)
  recorded_at: '2026-11-21T18:16:35.413Z'
id: TEST-ui-client-events-limit-network-and-cap
type: test
---
Unit tests of addressLimit with a fixed clock: thirty reports from thirty addresses of one /64 use up its allowance and the table holds one entry per network; an IPv4-mapped address shares its IPv4 address's allowance; 630 reports from many addresses in a minute take exactly 600, and one more is taken a minute later; and one address past its own limit leaves the shared cap to the others.