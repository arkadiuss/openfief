---
title: Stored answers are graded again by the current rules with no model call
status: active
origin:
  kind: agent
  ref: 'Design session decisions (author''s delegate): a scorer version and offline re-scoring of the Ask box bench, with D-217, D-219, D-220 and D-221, 2026-11-15'
  recorded_at: '2026-11-15T15:38:49.191Z'
id: SCEN-askbench-rescores-offline
type: scenario
---
Given a report.json of the bench, when it is re-scored with -rescore and a directory, then no model is called and no key is needed, a report of the current schema is re-graded from what it holds and changes nothing when no rule changed, a report of schema 1 has its cards rebuilt by replaying its calls and grades as it did, a run whose replay does not end as the run did is kept as graded and counted, the Limits check of such a report is carried over as it was graded, A19's two real answers that give the card's change now pass while the percent rounded another way and the fall still fail, and the new report holds the new scorer version, the case counts, the verdict and the ranking of the new grades.