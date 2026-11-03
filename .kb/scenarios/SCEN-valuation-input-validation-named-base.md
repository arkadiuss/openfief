---
title: Invalid valuation inputs, a rate without a named base among them, are refused
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:10.502Z'
id: SCEN-valuation-input-validation-named-base
type: scenario
---
Given the input recorder, when it is asked to record a non-positive price or one without an instrument, an FX rate that is not positive, has no base or is quoted in its own currency, an unknown index reading, bond terms without an ISIN or series, or an unknown tax wrapper, then it fails with an invalid input error and stores and enqueues nothing (D-180).