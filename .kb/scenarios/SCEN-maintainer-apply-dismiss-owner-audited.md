---
title: The owner applies or dismisses, audited, and applying changes weights only
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:32.713Z'
id: SCEN-maintainer-apply-dismiss-owner-audited
type: scenario
---
Given a pending proposal, when the owner applies it then the model portfolio's weights change in one transaction with the audit entry and nothing is traded; a model portfolio changed since refuses it with an audit entry and the owner dismisses it; an unrecorded decision changes nothing; a decided proposal cannot be decided again; any other session is refused.