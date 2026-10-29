---
title: An install argument is a catalog name or a digest-pinned image
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T22:23:48.088Z'
id: SCEN-plugins-cli-install-source
type: scenario
---
Given a catalog that pins sample at one digest and lists two versions of twice, when the owner installs sample, a mirror of sample's digest, an image the catalog does not pin, an image by tag, an unknown name and twice, then sample and the mirror resolve to the catalog entry and install as instance sample, the unpinned image installs under its repository's last path component as unverified, the tag is refused as not pinned, the unknown name is refused, and twice is refused naming both images.