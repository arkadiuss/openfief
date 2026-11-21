---
title: Tooling Vulnerability Findings Gate
status: active
fact_kind: subject
subject_key: tooling.vulnerability_findings_gate
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T12:51:08.000Z'
id: FACT-SUBJECT-F827044D9718484C
type: fact
---
The failing of the vulnerability check when govulncheck reports a vulnerability or npm audit reports a high or critical advisory the allow file does not list before its expiry.