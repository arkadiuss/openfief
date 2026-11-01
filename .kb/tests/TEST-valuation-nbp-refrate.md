---
title: NBP's reference rate archive is read exactly, in order, with its byte order mark
status: active
sourceFile: internal/valuation/nbp/refrate_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T03:56:15.380Z'
id: TEST-valuation-nbp-refrate
type: test
verification_scope: integration
---
A hypert replay of static.nbp.pl's public interest rate archive recorded on 2026-10-29 gives its 96 reference rate changes in strictly increasing order, spot-checked against the recorded text (24.00% from 1998-02-26, 3.75% from 2026-03-05). A synthetic archive with a byte order mark, changes out of order and one that sets other rates only gives two sorted changes as fractions. A 404, non-XML, another document, a bad date, a non-numeric rate, two changes on one day or no reference rate fail.