---
title: The Ghostfolio plugin syncs the owner's Ghostfolio on a timer with only what it was granted
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T01:36:53.895Z'
id: SCEN-plugins-ghostfolio-plugin
type: scenario
---
Given an installed ghostfolio instance with config url https://ghostfolio.example.com and interval 15m and the owner's security token in the ghostfolio-credentials Secret, when it starts, then it reads the read API with its projected token for audience openfief-core, reaches Ghostfolio only through the egress proxy, syncs at once and then every 15 minutes, logging only counts; given a plain http URL, an interval under five minutes or an ISIN listed twice, then it refuses to start; given no token in the Secret, then each sync fails before sending anything.