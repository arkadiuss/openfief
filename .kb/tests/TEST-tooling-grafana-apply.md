---
title: grafana.test.sh applies twice against a fake curl, keeps the token out of every output and refuses what it must
status: active
sourceFile: scripts/tests/grafana.test.sh
verification_scope: integration
origin:
  kind: agent
  ref: 'Author request: the demo''s Grafana Cloud stack as code, git-tracked JSON plus an idempotent apply script (docs/dev/demo.md#grafana), 2026-11-20'
  recorded_at: '2026-11-20T12:32:33.072Z'
id: TEST-tooling-grafana-apply
type: test
---
Runs scripts/grafana/apply.sh with a fake curl that plays a stack: the first apply creates the folder, applies the dashboard and creates each rule of deploy/grafana, the second updates them and changes nothing, rules go out without provenance and with the bearer header, and the token is in no output or log; each bad address, a missing token, a refused token, a 500 and an unreachable stack fail, naming the status and no body. It runs scripts/grafana/check.sh on the committed files, which pass, and on copies broken one way each, which it refuses with the reason, and checks that apply stops on a tree the check refuses.