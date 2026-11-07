---
title: Terms the owner leaves out come from the published data
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:07.318Z'
id: SCEN-treasurybonds-terms-from-published-data
type: scenario
---
Given a lot entered without some or all of its series' terms, when the connector maps it, then the missing rates come from the published data with its attribution and the missing fee from the reviewed table, terms the owner entered are kept, and a series neither covers is refused for the owner to enter.