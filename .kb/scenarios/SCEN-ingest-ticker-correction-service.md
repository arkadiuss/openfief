---
title: The owner corrects a confirmed ticker's ISIN and every row is booked anew
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T07:00:22.582Z'
id: SCEN-ingest-ticker-correction-service
type: scenario
---
Given a ticker confirmed with a wrong ISIN whose rows are booked, when the owner corrects it through the TickerService, then its transactions are booked anew under the corrected ISIN, superseding the old versions, its holding is asserted of the corrected instrument and the old holding withdrawn; a ticker never confirmed is failed_precondition, an unknown one not_found, and an account the owner may not change is refused before anything is booked.