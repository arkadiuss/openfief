---
title: Undeclared bond terms or index levels raise one attention item per instance until a clean delivery
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T22:24:56.674Z'
id: SCEN-ingest-undeclared-deliveries-item-per-instance
type: scenario
---
Given a connector instance whose signed manifest doesn't declare bond series terms or market index levels, or whose grant doesn't approve them, when its delivery carries some, then they are dropped with one skipped ingest problem per kind, the rest is imported, and one attention item of kind bond-terms-or-index-levels-undeclared is raised for the instance with a detail naming only the kinds, so dropping them again changes nothing; when its next delivery drops nothing, the item is cleared, and the UI links it to the plugins page (D-179).