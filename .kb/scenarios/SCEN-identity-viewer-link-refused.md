---
title: An expired or spent link is refused and recorded, and nobody is created
status: active
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: SCEN-identity-viewer-link-refused
type: scenario
---
Using a link after its lifetime, or a second time, fails with a forbidden error naming expiry or use, stores no person, and appends an enrollment_link_used entry with outcome refused naming the code.