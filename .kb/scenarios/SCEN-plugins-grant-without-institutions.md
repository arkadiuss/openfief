---
title: A grant that names no institutions approves none
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T12:31:49.166Z'
id: SCEN-plugins-grant-without-institutions
type: scenario
---
Given a connector whose signed manifest declares an institution, when its PluginGrant names no institutions, then the operator approves none: the declared institution waits in status.pendingApproval and a version that never ran does not start; when the field is removed from the grant of a running version, the institution waits for approval again and status.effective, what the core accepts deliveries for, names none (D-185).