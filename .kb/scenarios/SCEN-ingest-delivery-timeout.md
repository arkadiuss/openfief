---
title: A delivery whose statement runs past the delivery timeout fails whole and is a failed attempt of its own class
status: active
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:02.898Z'
id: SCEN-ingest-delivery-timeout
type: scenario
---
Given a connector sync whose delivery transaction has a statement that runs past the delivery timeout, when the transaction is run, then it fails with ErrDeliveryTimeout, the rows its earlier statements wrote and its audit entry are rolled back, and the attempt is recorded as failed with the class delivery-timeout; a delivery whose statements finish in time commits as before.

## Context

D-18 makes a delivery all or nothing and D-198 keeps every attempt with its error class. The timeout is a SET LOCAL statement_timeout of the delivery's transaction (--sync-delivery-timeout), so it ends with the transaction and bounds each statement, not the transaction as a whole. The class delivery-timeout says the core's database was too slow, not the connector.

## Source

> DELIVERY TIMEOUT: the sync delivery transaction (all-or-nothing, confirmed per D-18) gets a statement timeout (SET LOCAL statement_timeout inside it; value configurable, sane default e.g. 5 min). If a delivery hits it, it fails and is reported as a failed sync attempt (D-198) with a clear reason, never split silently.

Source: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18