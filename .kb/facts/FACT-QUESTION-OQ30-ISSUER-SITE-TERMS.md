---
title: Do obligacjeskarbowe.pl's terms of use allow a connector to read series terms from the site automatically (D-97, D-8)?
status: active
tags:
  - knowledge-source:design-open-questions
  - review:open-question
text_ref: design-open-questions:OQ-30
origin:
  kind: import
  ref: design-open-questions:OQ-30
  recorded_at: '2026-10-29T09:50:21.191Z'
fact_kind: observation
id: FACT-QUESTION-OQ30-ISSUER-SITE-TERMS
type: fact
---
Open question OQ-30, recorded in docs/DESIGN.md section 4 (Open for M0). It exists because D-97 has the treasury bond connector read each series' terms from the issuer's site, obligacjeskarbowe.pl, only once its terms of use are confirmed to allow it, and D-8 ships only connectors that comply with the institution's terms. Until it is answered, the owner enters series terms with a lot (REQ-ingest-bond-terms-entered-with-lot), and REQ-treasurybonds-series-terms-from-issuer waits on it.
