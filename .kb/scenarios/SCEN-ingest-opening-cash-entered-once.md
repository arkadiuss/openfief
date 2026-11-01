---
title: The owner enters an account's opening cash once
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:35:30.084Z'
id: SCEN-ingest-opening-cash-entered-once
type: scenario
---
Given the owner entering the account's cash with the opening upload of a statement that holds none, when the core hands it to the connector, then it is booked as the account's opening cash on the go-live date; cash sent with a transactions upload or for an account already seeded is refused, cash in another currency than the account's or in fractions of its minor unit is refused by the connector, and a statement that holds its own cash refuses it.