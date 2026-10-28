---
title: Rows a connector cannot book become ingest problems, and the rest is imported
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:31.191Z'
id: SCEN-ingest-problems
type: scenario
---
Given a Flex statement that mixes bookable rows with an option trade, a cancelled trade, a trade with an invalid ISIN, broker interest, an unreadable ending cash and a statement for an account that feeds no ledger account, when it is mapped, then the bookable rows become transactions and assertions while each other row becomes one unsupported or skipped problem naming its section, place and IBKR ID with a reason, and mapping does not fail.