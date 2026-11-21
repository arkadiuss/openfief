---
title: Release Fetch Script
status: active
fact_kind: subject
subject_key: tooling.release_fetch
origin:
  kind: agent
  ref: the design session's exit criterion (a) of 2026.11.0-alpha.2, relayed by the coordinator, 2026-11-21
  recorded_at: '2026-11-21T12:00:00.000Z'
id: FACT-SUBJECT-EF73BDB8CE72FB39
type: fact
---
The script that fetches a published OpenFief release and verifies it as the install guide tells an owner to, without touching a cluster: it downloads the release's bundle, CLI and checksums, checks their signatures and the images' signatures, and names the release's demo-seed image by its verified digest. It does not install anything, and it is not the post-release smoke test, which also installs.