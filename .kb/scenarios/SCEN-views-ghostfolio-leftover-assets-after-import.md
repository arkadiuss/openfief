---
title: An activity on an asset an earlier version left is deleted after the import
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T15:07:53.265Z'
id: SCEN-views-ghostfolio-leftover-assets-after-import
type: scenario
---
Given an owned account whose activity sits on the bond asset an earlier version keyed and named by the series code alone, when the Ghostfolio view syncs, then it imports the activity onto the new asset, named by the code plus a suffix no earlier version used, and only then deletes the old activity; only a stale activity an import duplicates is deleted before the import.