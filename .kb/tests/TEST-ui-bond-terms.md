---
title: The instrument page shows a bond's terms with their sources and attribution, and the owner sets its fee
status: active
sourceFile: internal/views/ui/bondterms_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:24.592Z'
id: TEST-ui-bond-terms
type: test
---
TestBondTermsOnInstrumentPage renders a bond's instrument page over the fake core for a read and an owner session: rates as percentages in the space-comma format, each term's source label, the to-confirm pill, the attribution line, and the fee form for the owner only; a bond without stored terms says the next sync fills them in, and a listed instrument has no terms section. TestSetBondFee posts the fee form: an empty field confirms, a decimal comma overrides, the core's refusal shows with 409, a non-number, a malformed form and a read session are refused. TestBondFeeItem checks the Inbox title and instrument link of the item.