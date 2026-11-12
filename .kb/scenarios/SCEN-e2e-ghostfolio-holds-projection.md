---
title: After a sync, the Ghostfolio accounts the view created hold exactly the projection of the record, matching the read API
status: active
tags:
  - acceptance:p10-9
origin:
  kind: agent
  recorded_at: '2026-11-08T08:59:30.183Z'
id: SCEN-e2e-ghostfolio-holds-projection
type: scenario
---
Given the demo record in OpenFief, a fresh Ghostfolio 3.93.0 in the same cluster with a new user whose base currency is PLN, and the Ghostfolio view installed with the CLI with that user's security token and Ghostfolio's Service as its in-cluster service and no listings, when the view's first sync has run, then it reports every owned account created, every projected activity imported and every projected cash balance set, leaving out only transactions of listed instruments; comparing the projection of what the read API answers with what Ghostfolio holds plans no change; every activity's comment names a current transaction of its account; each treasury bond series' units in Ghostfolio equal the read API's holding; and each owned account's cash equals the read API's cash of its account in its currency. Runs on both kind and k3d.