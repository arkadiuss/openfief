---
title: Opening lots named by ticker wait and then join the seeding
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:35:36.743Z'
id: SCEN-ingest-ticker-opening-lots
type: scenario
---
Given opening lots named by ticker only, with the date each was first acquired, when they cross the connector contract and their ticker's ISIN is not confirmed, then they wait like any ticker entry; once the owner confirms it, they are added to the account's seeding one at a time on its go-live date; and an opening batch of a broker account already seeded leaves them out.