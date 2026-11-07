---
title: A sync fills in the terms a lot's series lacked when it was entered
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:31.748Z'
id: SCEN-treasurybonds-sync-fills-missing-terms
type: scenario
---
Given a lot that left every term to the connector, of a series the ministry's data doesn't list yet, when it is entered, then it is booked without terms; when a later sync carries the accounts' holdings and the stored terms, then the connector delivers the terms of each held series of its own instruments that has none stored and the data now lists, with the table's or the assumed fee, never terms for a series whose terms are stored; a series still unlisted waits, and a ministry outage delivers no terms but still syncs.