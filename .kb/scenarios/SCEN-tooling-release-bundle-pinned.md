---
title: The release bundle pins every image of the install to the given digest
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T13:50:28.009Z'
id: SCEN-tooling-release-bundle-pinned
type: scenario
---
Given the images a release pushed and signed, as repository@sha256 digests or images.yml's digest files, when the release renders its bundle, then the install file runs the operator, the egress proxy and the core only at those digests and the CRDs file holds the namespace and the CRDs; a missing or malformed digest, an image by tag or an image the install doesn't name fails the run and writes nothing, and the same inputs always give the same bytes.