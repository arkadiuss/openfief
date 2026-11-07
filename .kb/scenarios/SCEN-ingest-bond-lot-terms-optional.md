---
title: A bond lot may leave its terms to the connector
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:09.104Z'
id: SCEN-ingest-bond-lot-terms-optional
type: scenario
---
Given the Add bond lot form or the owner service, when the owner leaves both rates, the fee or all three empty, then the lot is accepted and handed to the connector saying so, and a rate without the margin, or the margin without the rate, is refused.