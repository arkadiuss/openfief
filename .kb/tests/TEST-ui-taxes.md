---
title: The settings page shows the yearly tax and classifies instruments; totals show the losses offset
status: active
sourceFile: internal/views/ui/taxes_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T18:33:11.802Z'
id: TEST-ui-taxes
type: test
---
The UI handler over fake core services renders the settings page's yearly taxes section with each class's average and exempt part, the taxed part, rate and tax, the instrument to classify, posts a classification and shows it chosen, shows a refusal, and renders the losses-offset line under the overview's accounts and on the account page.