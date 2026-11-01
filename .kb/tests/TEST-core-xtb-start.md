---
title: An XTB account starts from a current export's open positions through the real core
status: active
sourceFile: cmd/core/xtbstart_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-01T07:35:20.490Z'
id: TEST-core-xtb-start
type: test
---
Runs the real core against a disposable PostgreSQL with a passkey owner session, the demo registry and an in-process connector that maps synthetic exports with xtb.Mapper. It uploads a current export while the broker account is unmapped, maps it, uploads with 4300 PLN of opening cash (opening mode, the cash booked and the lot waiting), uploads again with the core choosing (transactions, the rows before go-live skipped), with cash (refused), in opening mode without cash (nothing booked) and with other cash (refused), confirms the ticker (the opening lot booked), uploads a later export (the two rows after go-live booked) and reads the account's transactions through the read API.