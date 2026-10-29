---
title: Bond lots and bond terms cross the connector contract as exact strings, and refusals stay apart
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T01:13:58.461Z'
id: SCEN-ingest-bond-lot-messages
type: scenario
---
Given lots to hand to a connector, when the core calls ImportBondLots, then each lot's series, quantity, date, rates and fee go as exact strings, the delivered bond terms are read only when their symbol is a series code, their rates plain decimals and their fee a PLN amount of zero or more, and the connector's INVALID_ARGUMENT, UNIMPLEMENTED and unavailability come back as three different errors.