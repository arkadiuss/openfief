---
title: D-9 includes uploading files in the UI, and bond lots are entered by hand, but D-2 says views never write back and D-7 says data enters only through connector ingest paths. Is the built-in UI more than a view (it forwards uploads and bond-lot forms to the connectors), or does each connector serve its own small form, with the UI only linking to it?
status: closed
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
Open question OQ-26, answered by the author in decision D-49 (docs/DESIGN.md): file uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.
