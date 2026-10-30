---
title: Bond and fee assets are keyed by instance, user and code, and old keys migrate
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T01:20:56.234Z'
id: SCEN-views-ghostfolio-asset-keys
type: scenario
---
Given a view instance and a Ghostfolio user, when bonds and fees are projected, then each bond's MANUAL asset symbol is the UUID version 5 of the instance's name, the user's ID and the series code, named by the code, and each fee asset's is the same of the currency, both pinned and different for another instance or user. Given an owned account holding a bond on the asset an earlier version keyed by the series code alone, when the view syncs against Ghostfolio 3.93.0, then that activity is deleted before the import and the bond is imported onto its new asset with its full price history, and a second sync changes nothing.