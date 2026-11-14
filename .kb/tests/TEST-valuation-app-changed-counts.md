---
title: The input recorder hands the store's count of FX rates, readings and levels back
status: active
sourceFile: internal/valuation/app/changed_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:01.436Z'
id: TEST-valuation-app-changed-counts
type: test
---
TestRecordFXRatesChangedReportsTheStoresCount, TestRecordIndexReadingsChangedReportsTheStoresCount and TestRecordIndexLevelsChangedReportsTheStoresCount check the count comes back after the same validation the plain recording makes, and that an invalid input is an error with no count.