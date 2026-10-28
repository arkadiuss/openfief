---
title: Plugin images are accepted only when pinned by a sha256 digest
status: active
sourceFile: internal/plugins/image_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:03.340Z'
id: TEST-plugins-image
type: test
---
Table tests call ParseImageRef with synthetic registry.example.com references and fake digests. They accept a registry, a registry with a port, no registry and every legal path separator, and check Repository, Digest and String. They reject a tag only, no tag and no digest, a tag together with a digest, a short, upper-case or non-hex digest, a sha512 digest, upper-case paths, empty components, bad separators, a non-numeric registry port and two digests. A false pass would be a mutable tag accepted; the tag-only and tag-plus-digest cases guard it.