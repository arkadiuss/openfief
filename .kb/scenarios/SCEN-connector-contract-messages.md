---
title: The core and a connector exchange a batch over the connector contract
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:00:40.485Z'
id: SCEN-connector-contract-messages
type: scenario
---
Given a connector instance serving openfief.connector.v1.ConnectorService, when the core calls Sync or ImportFile with a batch ID it generated, then the connector answers with a delivery that echoes the batch ID, with amounts, prices, rates and quantities as exact decimal strings, dates as YYYY-MM-DD and currencies as ISO 4217 codes, and for a sync the watermark the next sync continues from.