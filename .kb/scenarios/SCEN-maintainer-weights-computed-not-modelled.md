---
title: Weights and drift are exact decimals computed by OpenFief
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:26.674Z'
id: SCEN-maintainer-weights-computed-not-modelled
type: scenario
---
Given the owner's valued holdings and a model portfolio, then each instrument target's weight is its value over the whole portfolio, rounded half-even to two places, its drift the weight minus the target, an index has neither and nothing worth nothing is weighed; a holding without a price makes the whole valuation unusable.