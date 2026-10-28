---
title: The read API's log and headers carry no financial data
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:43:25.513Z'
id: SCEN-views-api-logs
type: scenario
---
Given calls that succeed, name unknown accounts, fail in the store or carry an uncheckable token, when the read API logs them, then each log line names only the request ID, method, status and the step that failed, never the path, account ID, institution, instrument, amount, currency, token or error text; every response, errors included, carries Cache-Control no-store, nosniff, frame denial, a restrictive Content-Security-Policy, no-referrer and its X-Request-ID.