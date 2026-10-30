---
title: A connector declares cross-check prices and statement files in its manifest
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T04:10:55.100Z'
id: SCEN-plugins-manifest-declarations
type: scenario
---
Given ibkr-flex's manifest label with crossCheckPrices and statementFiles true, when it is parsed, then the manifest carries both and differs from one without them; and given a Plugin resource whose running version's manifest declares cross-check prices, when the core's discovery reads it, then the instance declares cross-check prices and not statement files, and a Plugin that isn't there declares nothing.