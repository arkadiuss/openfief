---
title: Reports from one IPv6 /64 share one allowance, and every address together is capped
status: active
origin:
  kind: agent
  ref: Security review of the new surfaces before the alpha.2 cut, /client-events (2026-11-21)
  recorded_at: '2026-11-21T18:16:36.670Z'
id: SCEN-ui-client-events-limit-network-and-cap
type: scenario
---
Given the UI's limit on client event reports, when one host sends 30 reports in a minute each from another address of the same IPv6 /64, then the next report from any address of that /64 is refused while an address of another /64 is still taken, and the table holds one entry per network; when the same IPv4 address is written as IPv4-mapped IPv6 it shares that address's allowance; when 630 reports come in a minute from many addresses, 600 are taken and the rest refused until the minute ends; and one address that keeps reporting past its own limit leaves the cap to every other address.