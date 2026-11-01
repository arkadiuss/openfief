---
title: The owner picks the charts' start in the page's URL and goes back to the default with a link
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T13:01:31.806Z'
id: SCEN-ui-chart-start-view-parameter
type: scenario
---
Given a session on the overview or an account's page, when the owner submits the Charts start on field, then the page reloads with start in its URL, both time charts start there, and a link without start leads back to the default; nothing is stored, so another browser or a reload without start shows the default.