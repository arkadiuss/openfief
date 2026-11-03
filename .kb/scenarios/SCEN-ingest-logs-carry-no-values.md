---
title: A failed sync or upload is logged without the values its error carries
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T01:05:18.855Z'
id: SCEN-ingest-logs-carry-no-values
type: scenario
---
Given a sync or a statement upload that fails with an error carrying an amount, an ISIN and an account number, such as an invalid transaction, an untranslatable delivery or a store failure quoting a row, when the scheduler or the statement service logs it, then the line names the failure's known kind, or only its Go type, and holds none of the values; a refused transaction's own error names what is wrong, never the amount or quantity; an invalid transaction or delivery in an upload is answered to the owner, not logged.