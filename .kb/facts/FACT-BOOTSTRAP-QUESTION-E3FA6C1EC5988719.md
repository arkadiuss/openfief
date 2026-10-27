---
title: 'Verify the Ghostfolio API details D-4 relies on, which are from memory or third-party summaries: creating accounts, listing activities per account, deleting activities, an import dry-run flag, the `MANUAL` scraper fields, and whether the import endpoint is meant to be stable.'
status: active
fact_kind: observation
text_ref: design-open-questions:OQ-23
tags:
  - bootstrap
  - knowledge-source:design-open-questions
  - review:open-question
origin:
  kind: agent
  recorded_at: '2026-10-27T20:13:10.947Z'
id: FACT-BOOTSTRAP-QUESTION-E3FA6C1EC5988719
type: fact
---
Open question OQ-23, recorded in docs/DESIGN.md section 8 and listed in section 10.3 as still open before M0 can be built. It exists because D-4's reconcile loop depends on Ghostfolio API calls known only from memory or third-party summaries (R-11); the Ghostfolio documentation site was unreachable from the review session.
