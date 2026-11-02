---
title: A pull request that touches the UI runs its browser checks
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T12:07:02.312Z'
id: SCEN-tooling-ci-ui-browser-checks
type: scenario
---
Given a pull request that changes only internal/views/ui/web/static/style.css, when task ci:plan decides which jobs run, then it prints ui=true and the ui-browser job runs task ui:browser. Given a pull request that changes only Go code the UI doesn't render, such as internal/ledger, or only docs, then it prints ui=false and the job is skipped. Given any push to main, it prints ui=true whatever changed. Assumes the UI reaches other packages only through internal/views, which a push to main covers when it doesn't.