---
title: A version names a catalog entry of the instance's own plugin
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T17:03:57.218Z'
id: SCEN-plugins-upgrade-version-resolution
type: scenario
---
Given an instance running a catalog-pinned image, when the owner upgrades it to version 2026.12.0, then the entry of that version of the same catalog plugin is chosen even if another plugin has that version too; given a version no entry of that plugin has, the upgrade is refused; given an image reference, it is used as for an install.