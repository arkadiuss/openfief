---
title: A quote the connector answers is checked, cut and cleaned before anything else sees it
status: active
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: SCEN-ingest-market-quote-answer-is-data
type: scenario
---
Given a connector that answers a quote, when its name holds line breaks, an instruction or more than 200 characters, then the core keeps one short line of printable text; when a close is not a plain positive amount, is out of date order, is dated after the period or is in another currency than its listing, or the symbol or ISIN is invalid, then the answer is an invalid message and no quote. Assumes the answer comes through the connector contract.
