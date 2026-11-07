---
title: A bond's instrument page shows its terms with their sources
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:34.214Z'
id: SCEN-ui-instrument-bond-terms
type: scenario
---
Given a treasury bond's instrument page, when any reader opens it, then its stored terms are shown in the reader's number format, each marked as published, from the fee table, assumed or entered by the owner, with the line naming the ministry, the dataset, CC BY 4.0, the days made and fetched and the processing by OpenFief under published ones; the owner gets the fee form, and a bond without terms says the next sync fills them in.