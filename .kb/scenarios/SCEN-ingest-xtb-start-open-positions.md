---
title: An XTB account starts from a current export's open positions
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:35:27.787Z'
id: SCEN-ingest-xtb-start-open-positions
type: scenario
---
Given a current xStation export uploaded in opening mode for an account not yet started, when the xtb connector maps it, then each open stock or ETF position quoted in the account's currency becomes an opening lot on the report's date, first acquired on its open date at its volume times its open price, a position quoted in another currency and each CFD are ingest problems, nothing of Cash Operations is booked, and the same export uploaded again changes nothing.