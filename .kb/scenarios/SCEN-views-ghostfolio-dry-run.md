---
title: Every import is checked with a dry run first
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:30.162Z'
id: SCEN-views-ghostfolio-dry-run
type: scenario
---
Given activities to import, when the view syncs, then it sends each batch of at most 500 activities with dryRun=true first, carrying the assets the batch uses, and imports it only when the preview shows each activity in the same account under the same symbol and none skipped as a duplicate; when the preview shows a duplicate or the owner's own asset under the bond's name, then the sync fails with ErrPreviewMismatch and nothing is imported or deleted.