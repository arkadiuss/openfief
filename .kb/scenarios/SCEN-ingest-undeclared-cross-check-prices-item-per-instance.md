---
title: Undeclared cross-check prices raise one attention item per instance until a clean delivery
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T13:17:34.976Z'
id: SCEN-ingest-undeclared-cross-check-prices-item-per-instance
type: scenario
---
Given a connector instance whose manifest doesn't declare cross-check prices, when its delivery carries some, then they are dropped, the rest is imported, and one attention item of kind cross-check-prices-undeclared is raised for the instance with a fixed detail, so dropping them again changes nothing; when its next delivery carries none, or only declared ones, the item is cleared, and the UI links it to the plugins page (D-141).