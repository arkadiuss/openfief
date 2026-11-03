---
title: The portfolio registry checks kinds and sleeves, ensures the defaults, places accounts in one sleeve and deletes only what nothing uses
status: active
sourceFile: internal/portfolio/app/registry_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T11:48:49.855Z'
id: TEST-portfolio-app-registry
type: test
---
Table tests drive portfolio/app.Registry over an in-memory RegistryStore with the Postgres store's rules and a fake ledger that holds acc-1 and acc-2. They check that SaveSleeve and SaveKind apply portfolio.CheckSleeve and CheckKind (role from the kind, the override, unknown kinds, a default role change refused), that EnsureDefaults stores the built-in kinds and, only on a fresh install, the default sleeves, that PlaceAccount puts an account the ledger holds into one stored sleeve and refuses an unknown account or sleeve, and that a sleeve with accounts and a kind in use are not deleted until the account leaves.