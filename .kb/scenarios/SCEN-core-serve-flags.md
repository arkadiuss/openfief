---
title: The core reads its serve flags, both ports' addresses included, with defaults and refuses bad ones
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T21:48:07.497Z'
id: SCEN-core-serve-flags
type: scenario
---
Given no flags, when serve is parsed, then the browser port listens on :8080, the plugin port on :8081 and the reporting currency is the pack's PLN; given --listen, --plugin-listen and the other flags, then they are taken; given a bad value or the removed --bond-basis flag, then serve refuses to start.