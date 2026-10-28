---
title: NBP table A rates are parsed exactly from the JSON number text
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:57:00.797Z'
id: SCEN-valuation-nbp-decimal-exact
type: scenario
---
Given an NBP table A response whose mid rates are JSON numbers such as 4.10120000000000000001 or 2.54E-4, when the adapter maps it, then each rate equals the exact decimal written in the response (no float rounding), dated on the table's effective date, per one unit of the currency and with source nbp-a; given a response that is not a valid table A (other table, bad date, bad code, non-positive or non-numeric rate), then the adapter fails with a malformed table error.