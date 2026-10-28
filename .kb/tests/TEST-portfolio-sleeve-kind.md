---
title: Sleeve kinds validate, keep their default role, and drive sleeve roles and install defaults
status: active
sourceFile: internal/portfolio/kind_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T09:34:18.315Z'
id: TEST-portfolio-sleeve-kind
type: test
---
Table tests over the pure portfolio functions. Kind.Validate accepts lowercase hyphenated keys up to 64 characters and rejects empty, uppercase, spaced, leading, trailing or doubled hyphen keys, a missing name and a missing or unknown default role. DefaultKinds is checked for the six built-in keys with their roles, key order and a fresh slice per call. CheckKind accepts a new kind and a rename and refuses changing a stored kind's default role either way. CheckSleeve fills a missing role from the kind, allows defensive in an offensive kind, allows offensive in a defensive kind only with OffensiveOverride, and refuses a stray override, a missing or unstored kind and an invalid sleeve. DefaultSleeves is compared in full and must pass CheckSleeve unchanged; MissingDefaults returns everything on a fresh install, nothing when all is stored, and keeps a renamed kind or a sleeve moved to another kind. KindBySleeve maps sleeve IDs to kinds.