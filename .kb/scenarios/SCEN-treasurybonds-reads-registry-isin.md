---
title: The connector reads each series' registry ISIN from the ministry's data
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:24.671Z'
id: SCEN-treasurybonds-reads-registry-isin
type: scenario
---
Given the ministry's workbook lists a series' code with a valid Polish ISIN in its Kod ISIN column, with or without its rates yet, when the connector reads it, then the series' registry ISIN is that one; a cell that is no valid Polish ISIN, or is a provisional one, gives none, and one ISIN listed for two series refuses the workbook.