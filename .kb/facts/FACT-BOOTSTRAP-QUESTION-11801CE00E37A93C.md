---
title: D-9 includes uploading files in the UI, and bond lots are entered by hand, but D-2 says views never write back and D-7 says data enters only through connector ingest paths. Is the built-in UI more than a view (it forwards uploads and bond-lot forms to the connectors), or does each connector serve its own small form, with the UI only linking to it?
status: active
fact_kind: observation
text_ref: design-open-questions:OQ-26
tags:
  - bootstrap
  - knowledge-source:design-open-questions
  - review:open-question
origin:
  kind: agent
  recorded_at: '2026-10-27T20:13:11.998Z'
id: FACT-BOOTSTRAP-QUESTION-11801CE00E37A93C
type: fact
---
Open question OQ-26, recorded in docs/DESIGN.md section 8 and listed in section 10.3 as still open before M0 can be built. It exists because D-9 has XTB files uploaded in the UI and bond lots entered by hand, while D-2 makes views read-only and D-7 routes all data through connector ingest paths. The author has not yet decided where those forms live.
