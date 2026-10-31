---
title: A transfer's comment is never kept
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T12:54:57.656Z'
id: SCEN-ingest-xtb-no-transfer-comments
type: scenario
---
Given deposits and withdrawals whose comments name a sending bank and the XTB account, when the export is mapped, delivered and imported, then no part of those comments is in the mapping, the delivery or any stored column, and the account number names only the broker account.