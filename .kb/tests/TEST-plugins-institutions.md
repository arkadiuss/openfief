---
title: A connector's manifest label declares the institutions it books for
status: active
sourceFile: internal/plugins/institutions_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T22:19:38.629Z'
id: TEST-plugins-institutions
type: test
---
TestManifestInstitutions parses synthetic manifest labels in process: a connector's institutions are read in order, an institution that isn't a lower-case DNS label, one named twice and one declared by a view are refused with their field errors, and two manifests that declare the same institutions in another order are equal while other institutions make them differ, which is what makes the operator refuse a Plugin whose institutions are not its image's signed label.