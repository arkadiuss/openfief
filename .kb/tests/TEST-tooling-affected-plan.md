---
title: The hook's plan for each kind of change
status: active
sourceFile: tools/affected/plan_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T03:32:17.150Z'
id: TEST-tooling-affected-plan
type: test
verification_scope: unit
---
Table tests build a synthetic package graph (a shared kernel, a ledger on it, a valuation package whose tests import the ledger, a command, a plugin on the SDK module, the operator whose tests read deploy/crds, the UI embedding a file, the e2e suite) and map change sets to the exact task targets and variables: docs-only and kibi-only changes run docs:check, docs:build and kibi:staged and no Go step; one plugin checks only that plugin; a shared package checks every importer and lints the packages whose tests import it; an SDK change reaches the root packages importing it; embedded files, test data, generated manifests, workflows, deletions and plugin manifests select their steps; go.mod, an unknown root file, a Go file in no package and a file no package reads select the full check. A second test checks the per-package reasons and the skipped steps. A false pass would be a step list compared loosely: the whole list is compared with go-cmp.