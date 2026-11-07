---
title: The owner confirms or overrides a series' early-redemption fee
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:29.250Z'
id: SCEN-valuation-owner-sets-bond-fee
type: scenario
---
Given a series' stored terms, when the owner sends no fee, then an assumed fee becomes the owner's and nothing is revalued, and a fee that is not assumed is refused as a failed precondition; when the owner sends a fee, it replaces the series' fee whatever it came from and every account that held the series is revalued if the amount changed. A bond without terms is not found, a malformed ISIN or fee is invalid, and only the owner may change it, from the instrument's page.