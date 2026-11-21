---
title: The vulnerability check must fail when govulncheck reports a vulnerability, or when npm audit reports a high or critical advisory that scripts/npm-audit-allow.txt does not list before its expiry; an entry past its expiry fails it too.
status: open
semantic_text: The vulnerability check must fail when govulncheck reports a vulnerability, or when npm audit reports a high or critical advisory that scripts/npm-audit-allow.txt does not list before its expiry; an entry past its expiry fails it too.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: eddad86d609ad5738638d3b97ef0250242255579fda2530c783a0b8c6f3171d2
semantic_inventory:
  - claim_key: CLAIM-7A2F389FF1E24834
    claim_text: The vulnerability check must fail when govulncheck reports a vulnerability, or when npm audit reports a high or critical advisory that scripts/npm-audit-allow.txt does not list before its expiry
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 194
    subjects:
      - tooling.vulnerability_findings_gate
  - claim_key: CLAIM-BDC5D9D8953E20BA
    claim_text: an entry past its expiry fails it too
    role: descriptive
    status: missing
    span:
      start: 196
      end: 233
logic_claims:
  - CLAIM-7A2F389FF1E24834
  - CLAIM-BDC5D9D8953E20BA
text_ref: docs/dev/sdlc.md
origin:
  kind: agent
  ref: 'Author''s answer 2026-11-18 ("Try 2 if not then 1"): override what has a fix, allowlist with expiry what has none'
  recorded_at: '2026-11-18T06:03:05.634Z'
id: REQ-tooling-vuln-findings-fail-v2
type: req
---
The vulnerability check must fail when govulncheck reports a vulnerability, or when npm audit reports a high or critical advisory that scripts/npm-audit-allow.txt does not list before its expiry; an entry past its expiry fails it too.

## Context

Part of the local pipeline the author asked for: tests, lint and vulnerability checks run on the pre-commit hook and in CI (docs/dev/sdlc.md). On 2026-11-18 a newly published picomatch ReDoS advisory (GHSA-fc5q-39xc-vjpf, high) blocked every commit through the npm dev tooling only (kibi and patch-package; package.json has devDependencies only and nothing ships in an image), with no fixed release of picomatch yet. The author chose, through the orchestrating agent: override what has a fix (minimatch 11.0.9, hono 4.13.19, @hono/node-server 2.1.9 in package.json overrides), and for what has none, a named advisory on an allowlist with a short expiry, after which the check fails again so someone looks. Any other high or critical advisory, and govulncheck's findings, still fail.

## Why this changed

REQ-tooling-vuln-findings-fail said any npm audit finding fails the check. A dev-only advisory with no fixed release (picomatch, GHSA-fc5q-39xc-vjpf) would have blocked every commit indefinitely, so the author allowed named advisories with an expiry; this requirement supersedes it.

## Source

> If a vulnerability database is unreachable, `task vuln` warns locally and fails in CI (`CI=true`). Findings always fail.

docs/dev/sdlc.md#tooling

> Try 2 if not then 1

The author's answer, 2026-11-18, to: (1) an expiring allowlist of the named dev-only advisories, or (2) first try forcing minimatch@11 under ts-morph and test kibi; picomatch has no fix, so it needs an allowlist entry either way.
