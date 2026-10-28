---
title: The EODHD client sends the owner's key as api_token and never returns it
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:12.183Z'
id: SCEN-valuation-eodhd-request
type: scenario
---
Given the owner's API key and a caller-supplied HTTP client, when closes of SYNT.XETRA are asked for from 2025-01-02 to 2025-01-07, then one GET goes to eodhd.com/api/eod/SYNT.XETRA with api_token, fmt json, period d, order a, from and to, and User-Agent openfief; given a zero from, then from is omitted; given 404, then the error is ErrNotFound; given 401 or 402, then ErrUnexpectedStatus; given a transport failure, then the error holds neither the key nor the URL.