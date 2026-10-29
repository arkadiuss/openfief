---
title: Each installed plugin shows one of four trust badges with its exact signer
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T09:49:45.242Z'
id: SCEN-ui-plugin-trust-badges
type: scenario
---
Given installed plugins with each signature trust the read API reports, when the plugins page or the settings page lists them, then a catalog-reviewed plugin shows Verified with a check, an owner-signed one Yours, one signed by OpenFief's CI, a catalog publisher or another trusted signer but not in the catalog Unreviewed, and a refused, unchecked or unknown one Blocked with a cross, each with the full signature trust and signer under the badge.