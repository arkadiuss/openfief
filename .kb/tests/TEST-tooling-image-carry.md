---
title: A carried image keeps a digest the images workflow signed
status: active
sourceFile: scripts/tests/image-carry.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-30T03:32:41.839Z'
id: TEST-tooling-image-carry
type: test
---
Runs scripts/image-carry.sh with a fake crane and a fake cosign over a fake registry directory: a signed digest gets the sha-<commit> and main tags and the tag is verified after tagging; an unsigned digest is refused before any tag is written; a tag that resolves to another digest after tagging fails; malformed image, digest, commit or repository arguments are refused before any call.