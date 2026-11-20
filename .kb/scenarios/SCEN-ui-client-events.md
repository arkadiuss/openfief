---
title: The pages' client events are validated, capped, logged and counted, and can be turned off
status: active
origin:
  kind: agent
  ref: D-227, 2026-11-18
  recorded_at: '2026-11-18T21:55:27.594Z'
id: SCEN-ui-client-events
type: scenario
---
Given the UI, when a page posts a report of Turnstile's states, a script error and a failed request, then every event that fits the fixed vocabulary is logged without an address and counted by kind, browser family and code, one that does not fit is dropped, a message is kept only from the UI's own scripts and only as plain words, a report over 4 KiB or 20 events or past 30 a minute from one address is refused or cut, and in the browser each state of the fake widget's stalled and visible paths is reported with its time and path and the payload holds nothing typed and no token; with reporting off the page loads no reporting script and the endpoint answers 204 and discards.