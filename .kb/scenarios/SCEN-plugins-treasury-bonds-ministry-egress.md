---
title: The connector may reach gov.pl and dane.gov.pl only
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:12.667Z'
id: SCEN-plugins-treasury-bonds-ministry-egress
type: scenario
---
Given the treasury-bonds manifest, when it is read, then it asks for egress to www.gov.pl and api.dane.gov.pl and for https-only base URL fields for them, and names no obligacjeskarbowe.pl; a mirror's host is asked for in the grant.