---
title: The egress audit fails on an allowed request off its instance's grant and on a log line holding more than who, where and the decision
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T03:30:46.979Z'
id: SCEN-plugins-egress-audit-log
type: scenario
---
Given the egress proxy's decision lines and each instance's approved hosts, when the audit runs, then allowed requests to approved hosts and refusals pass and are counted by instance and host, while an allowed request to a host off the asking instance's grant, a field beyond who asked, where to, the decision and its timing, a requested host that is not a bare host and port, and a line without its instance or decision each fail it, reported by line id, instance and field name and never by a value.