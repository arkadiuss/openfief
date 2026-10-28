---
title: OpenFief documentation
description: A self-hosted portfolio tracker that grows into an orchestrator for AI investment agents, each holding a revocable capital lease with hard limits enforced outside the model.
---

OpenFief is in its first milestone, M0: a tracker for Polish investors with imports from their brokers and the treasury bond registry, run on Kubernetes, with plugins isolated from each other and from the internet. Agents come later. These pages describe what is built today, why it is built that way, and how to check it.

## Start here

- [Overview](../../docs/OVERVIEW.md) introduces the idea on one page.
- [Architecture](../../docs/architecture.md) shows how the parts fit together, with diagrams.
- [Status](../.gen/content/status.md) lists every decision with the requirements, code and tests behind it: what is built and what is verified.
- [Concepts](../../docs/concepts/) explains each part as built, one page per area.

## The design and its vocabulary

- [Design](../../docs/DESIGN.md) holds the decided design: ground rules and decisions, each with a stable ID such as D-18.
- [Glossary](../../docs/glossary.md) defines every domain term and abbreviation.
- [Proposals](../../docs/proposals/) are ideas for later milestones; nothing there is decided.
- [Research](../../docs/research.md) holds the external facts the design relies on, with dates and sources.

## How these pages stay true

Reference pages are generated from the repository on every build, so they cannot drift. Each hand-written page names the code it covers, and a change to that code must update the page or say why not. Every diagram is rendered and every internal link resolved at build time, so a broken one fails the build. [Documentation](../../docs/dev/docs.md) explains the rules.
