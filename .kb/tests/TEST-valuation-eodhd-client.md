---
title: The EODHD client searches by ISIN and fetches closes exactly, never leaking the API key
status: active
sourceFile: internal/valuation/eodhd/client_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T21:36:58.939Z'
id: TEST-valuation-eodhd-client
type: test
verification_scope: integration
---
hypert replays SYNTHETIC recordings hand-written in EODHD's documented /api/search and /api/eod shapes (record switch off, api_token sanitized in stored and replayed requests): a search for XX00000SYNT3 gives its Xetra, LSE and Frankfurt listings without the fuzzy match lacking an ISIN; an empty search; four closes of SYNT.XETRA in January 2025; the whole history of FAKE20.INDX when no from is given; and 404 as ErrNotFound. Replay validates path, query including fmt, period, order, from and to, method, scheme and headers including User-Agent openfief. Stub transports check 20-digit and exponent closes survive exactly and sorted, null and non-positive closes are dropped, 401, 402 and 500 are ErrUnexpectedStatus, bad bodies, dates and text closes are ErrMalformedResponse, and a transport error carries neither the key nor the request URL.