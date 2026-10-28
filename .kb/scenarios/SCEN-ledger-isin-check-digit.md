---
title: An ISIN with a wrong check digit is rejected
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:23:53.126Z'
id: SCEN-ledger-isin-check-digit
type: scenario
---
Given the twelve-character text US0378331006, which has the shape of an ISIN but whose last digit is not its Luhn check digit, when it is parsed as an ISIN, then it is rejected as an invalid ISIN; and given an instrument with an ISIN but no venue symbol, when it is validated, then it is rejected. Assumes ISINs arrive upper case from the connector.