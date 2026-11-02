---
title: Before the first beta a contract breaks only in a commit marked breaking
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T20:54:36.406Z'
id: SCEN-tooling-breaking-before-first-beta
type: scenario
---
Given a repository whose tags name no beta or stable release, when a change breaks a contract with no exceptions entry, then the working-tree check passes and says what it allowed, while the commit-msg hook and the check against the base refuse the commit until it carries `!` or a BREAKING CHANGE footer; once a beta tag exists the same change fails unless an author decision's new entries excuse it. Exceptions entries are checked either way.