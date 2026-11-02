---
title: XTB's free-funds interest is cash income net of its tax
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T05:51:19.937Z'
id: SCEN-ingest-xtb-free-funds-interest
type: scenario
---
Given an xStation export with a Free funds interest row and the Free funds interest tax rows of its date, when it is mapped, then the interest is cash income of its amount with the summed tax withheld, keyed by its operation ID, and a tax row with no single interest row of its date, or a refund, is an ingest problem.