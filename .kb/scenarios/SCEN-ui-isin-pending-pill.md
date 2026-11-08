---
title: A pending pill shows next to a provisional ISIN
status: active
origin:
  kind: agent
  recorded_at: '2026-11-08T02:00:39.947Z'
id: SCEN-ui-isin-pending-pill
type: scenario
---
Given a bond account holds a series the Ministry of Finance's data doesn't list yet, under its provisional ISIN, when the owner opens the account's holdings or open lots, after entering a lot too, the holdings behind its allocation bar or its instrument page, then a neutral pending pill sits next to the series once, whose tooltip says it moves to the registry's ISIN when the data lists the series; an instrument under a registry ISIN has none.