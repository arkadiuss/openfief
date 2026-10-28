---
title: The NBP adapter fetches table A exactly and treats a range without tables as empty
status: active
sourceFile: internal/valuation/nbp/client_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:56:57.635Z'
id: TEST-valuation-nbp-client
type: test
---
hypert replays two committed recordings of the public api.nbp.pl (record switch off): table A for 2 to 7 January 2025 gives 33 rates on each of the three business days with spot-checked rates such as USD 4.1219 and IDR 0.00025276 on their dates, and the 4 to 6 January weekend plus Epiphany gives NBP's 404, which must be an empty result. Replay validates path, query, method, scheme and headers including the adapter's own User-Agent. A stub transport checks that 20-digit and exponent-form rates survive exactly, that the User-Agent is sent, and that over-93-day ranges, reversed ranges, non-200/404 statuses, non-JSON bodies, another table, a bad date, a bad currency code, a zero rate and a text rate are rejected with typed errors.