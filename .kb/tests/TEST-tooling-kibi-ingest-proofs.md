---
title: kibi-ingest-proofs.sh hands each end-to-end proof run to kibi, oldest first, and stops at a refusal
status: active
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T12:45:29.368Z'
id: TEST-tooling-kibi-ingest-proofs
type: test
---
scripts/tests/kibi-ingest-proofs.test.sh runs scripts/kibi-ingest-proofs.sh in a mktemp directory with a fake kibi on PATH that logs each ingest-proof call and its input. With no artifact the script fails without calling kibi. With a k3d artifact that finished after a kind artifact, kibi gets the kind one first, each whole with its own code_snapshot as the snapshot. When kibi refuses the first artifact the script fails, prints the refusal and never ingests the second. A false pass would be ingesting by path order; the k3d artifact sorts first by path but finished later.