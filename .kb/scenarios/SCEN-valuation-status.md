---
title: The status shows valued-through, pending and running
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:56:36.610Z'
id: SCEN-valuation-status
type: scenario
---
Given acc-1 valued through 2026-03-03, then it is neither pending nor running; when revaluations of acc-1 and acc-2 are enqueued, then both are pending; when acc-2's job runs, then acc-2 is running and not pending; when it fails and waits for a retry, then it is pending again; an unknown account has no date and neither flag.