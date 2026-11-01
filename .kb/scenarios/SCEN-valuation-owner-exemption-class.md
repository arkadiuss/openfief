---
title: The owner classifies an instrument
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T18:33:31.373Z'
id: SCEN-valuation-owner-exemption-class
type: scenario
---
Given an owner session, when the owner sets an offered class for an ISIN, then it is stored with the instrument, replacing an earlier one, revalues nothing, and another class, a bad ISIN or another session is refused.