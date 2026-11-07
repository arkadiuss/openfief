---
title: The connector fills in the terms the owner left out
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:07.318Z'
id: SCEN-treasurybonds-terms-from-published-data
type: scenario
---
Given a lot entered without some or all of its series' terms, when the connector maps it, then the missing rates come from the published data with its attribution and the missing fee from the reviewed table, assumed past the reviewed month, terms the owner entered are kept, a lot that left everything to a series the data doesn't list yet is booked without terms, and one whose fee the owner entered is refused for the owner to enter the rates.