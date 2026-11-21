---
title: Release Upgrade In Place
status: active
fact_kind: subject
subject_key: release.upgrade_in_place
origin:
  kind: agent
  ref: the design session's exit criterion (a) of 2026.11.0-alpha.2, relayed by the coordinator, 2026-11-21
  recorded_at: '2026-11-21T12:00:00.000Z'
id: FACT-SUBJECT-AC56307FAB0C2D30
type: fact
---
An upgrade in place of an OpenFief install from one published release to a later one: the owner applies the later release's files, with the install guide's commands and in its order, over the running install, which keeps its database. It covers the migration of what is stored, the one-time household-grant cutover and SpiceDB's credentials Job at the first start of the new core. It does not cover a fresh install, or upgrading a plugin instance.