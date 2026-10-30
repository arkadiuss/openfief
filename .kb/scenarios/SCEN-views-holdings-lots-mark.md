---
title: Holdings and lots behind a removed connector instance are marked as no longer updated
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T00:04:54.774Z'
id: SCEN-views-holdings-lots-mark
type: scenario
---
Given a removed instance that bought units of an instrument, when holdings and lots are read, then the holding and the lot it opened are marked with that instance, a lot only cut by a disposal is marked when that disposal's instance is removed, and nothing is marked while the installed instances are unknown.