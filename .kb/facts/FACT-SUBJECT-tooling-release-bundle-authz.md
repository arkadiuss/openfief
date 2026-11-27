---
title: Tooling Release Bundle Authz
status: active
fact_kind: subject
subject_key: tooling.release_bundle_authz
origin:
  kind: agent
  recorded_at: '2026-11-22T08:41:08.401Z'
id: FACT-SUBJECT-tooling-release-bundle-authz
type: fact
---
The --authz flag the core container of a release bundle's install file runs with, which decides whether SpiceDB or the policy answers authorization on an install made from it, and the bundle script's refusal of any other than --authz=spicedb. It is about what a release ships, not about the flag's handling in the core.