---
title: kibi-timed.sh prints a kibi command's duration and warns from 90 seconds
status: active
sourceFile: scripts/tests/kibi-timed.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-07T19:27:53.281Z'
id: TEST-tooling-kibi-timed
type: test
---
Runs scripts/kibi-timed.sh with a fake kibi and a fake clock whose `date +%s` answers come from a list, so no case waits: 30 s prints the duration and no warning, 89 s is not a warning yet, 90 s under GitHub Actions is a ::warning:: workflow command on stderr while stdout stays exactly kibi's, 95 s outside Actions warns in plain text, kibi's stderr, arguments and exit status pass through, and a missing command is a usage error.