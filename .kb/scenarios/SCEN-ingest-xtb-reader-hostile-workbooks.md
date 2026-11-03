---
title: Hostile XTB workbooks are refused within a bounded heap
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T07:17:57.660Z'
id: SCEN-ingest-xtb-reader-hostile-workbooks
type: scenario
---
Given an uploaded XTB export, when an attribute or text in it is longer than 256 KiB, even made of '>', when its zip has more than 1,000 entries, or when it lists two sheets on one part, then the reader refuses it before keeping anything large; when a sheet has many empty rows, none of them is kept; and the largest workbook the reader keeps grows the heap far less than the plugin pod's 256 MiB memory limit.