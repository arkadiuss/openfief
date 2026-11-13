---
title: On k3d and kind, the IBKR Flex, XTB and bond data path against the institution fixture balances, imports idempotently with versions, seeds opening lots with PLN cost and values at NBP rates and EODHD prices
status: active
sourceFile: test/e2e/datapath_test.go
verification_scope: end_to_end
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:28.912Z'
id: TEST-e2e-data-path
type: test
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e
  required_proofs:
    - symbol_id: SYM-e2e-datapath-test-test-data-path
      target: default
  success_policy: all_required_first_attempt
proof_bindings:
  - symbol_id: SYM-e2e-datapath-test-test-data-path
    target: default
    source_file: test/e2e/datapath_test.go
    native_id: test/e2e/datapath_test.go::TestDataPath
---
TestDataPath deploys the institution fixture as an HTTPS Service in its own namespace with a certificate from a CA made for the run, installs nbp-fx, eodhd, ibkr-flex, treasury-bonds and xtb with the openfief CLI built from the checkout, each signed with the run key, the network ones with their base URL on the fixture, the fixture Service as their in-cluster service and the run CA in their ca-bundle Secret. The owner signs in with a passkey in Chromium through localhost (owner.mjs) and, through the UI forms, creates an IBKR account mapped to the reported broker account, an XTB account and a bond account. The IBKR account is seeded from the opening statement, then synced through the week and its correction by changing the query ID in its Secret; the XTB account is seeded from an uploaded export, its ticker confirmed, and a later export uploaded twice; a fixed-rate bond lot is entered twice. It checks with core check ledger and the read API: a balanced whole log, unchanged logs after re-sync and re-upload, the corrected deposit as a second version with the first superseded, the opening lots with open dates and costs, net and gross PLN values and cash hand-computed from the fixture NBP rates and EODHD closes, and the price checks: IBKR mark 1.96% off flagged, 0.48% off matched. A false pass would need the hand-computed values to coincide with other rates, which the rates rising each weekday rule out. At the end it checks criterion 8's charts (charts_test.go): owner.mjs opens the overview's value chart, gross and all three values, its performance and its allocation by every dimension of every value, the IBKR, XTB and bond account pages and the three instrument pages, and reads what each Chart.js chart holds once app.js drew it and each allocation bar's value, share and width; each must equal what the read API answers the script token for the same days, worked out as the UI asks (test/e2e/chartcheck), retried up to five times while the record changes. Then the core's log over the spans the UI drew those pages in must show the ui plugin calling nothing outside the read API but the AuthService, ListStatementAccounts, ListConfirmedTickers and GetAskModel, with read API requests in each span, and the UI's NetworkPolicy must let it reach only the core's plugin port, with no proxy, egress token or allowed host. A false pass of the chart check would need the UI and the check to work the expected lines out the same wrong way. Last, criterion 19 (model_test.go): with the model portfolio off, never saved and then saved with the box unchecked, the read API must answer 404 and the overview's performance chart must draw the time-weighted return alone with no model portfolio section; the owner then saves two index targets in PLN, MODA at 60% and MODB at 40%, through the settings form and turns it on, an eodhd sync fetches the two synthetic series' levels from the fixture (test/e2e/modelseries), and the read API's model portfolio from December 3, 2025 and from March 6, 2026 must equal the hand calculation of test/e2e/modelcheck: each day's index across the yearly rebalance on December 31 to 12 places, the model's max drawdown, 0.35 from January 20 to February 16, MODA's 0.25 from January 15 to February 2 and MODB's 0.55 from December 12 to February 16, with their dates and values; the overview's performance chart must draw the model as a second line over the same days and the page must list the three drawdowns; turned off again, both show nothing. A false pass would need the backtest, the fixture and the hand table to share a mistake: the table is checked against portfolio.Backtest in the unit test, and the levels are chosen so the rebalanced and the unrebalanced basket, the earlier falls and the max falls all differ.