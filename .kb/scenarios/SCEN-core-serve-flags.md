---
title: The core reads its serve flags with defaults and refuses bad ones, the bond basis flag included
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T21:48:07.497Z'
id: SCEN-core-serve-flags
type: scenario
---
Given no flags, when the core parses serve, then it listens on :8080 with PLN reporting from the jurisdiction pack and the documented defaults; given valid flags, then they are taken; a bad currency, time zone, end of day, threshold, or the --bond-basis flag removed by D-77, makes it refuse to start.