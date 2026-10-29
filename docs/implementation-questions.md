---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## Blocking real use

1. **XTB import needs the export's real layout.**
   - *Problem:* D-9 says XTB data arrives as uploaded xStation export files, but the export's columns, and whether it separates CFDs from stocks, are unverified (OQ-22).
   - *Today:* no XTB parser exists.
   - *Needed:* a header row and two or three rows with made-up values, copied from a real export. The parser's fixtures will be synthetic.
   - *Known so far:* the author holds stocks and ETFs only at XTB (so D-24's CFD handling stays as written) and will send the headers from a real export. The export's language is not known yet. Public parsers disagree on the sheet names (R-19).

## Identity and the UI

2. **How the built-in UI formats amounts.**
   - *Problem:* the design system says amounts are formatted for the reader's locale, Polish being `5 000,00 PLN`, but no decision says which locale the UI uses: its pages are in English, and OpenFief is built for Polish investors first.
   - *Today:* every amount keeps its exact digits and its currency, with thousands separated by a narrow no-break space and a decimal point, such as `5 000.00 PLN`, whatever the browser's language. The charts' axes use Chart.js's own number format.
   - *Options:* keep one format; follow the browser's language (`Accept-Language`); or a setting next to the reporting currency.

## End-to-end tests

3. **How the end-to-end suite gets signed plugin images to the operator.**
    - *Problem:* the operator deploys only images whose signature it reads from their registry (D-56, D-57), and kind and k3d configure their registries differently (D-12).
    - *Today:* `task e2e:images` runs a `registry:3.0.0` container on the cluster's Docker network and pushes `nbp-fx` there; the suite signs it with a key made for the run, trusted through `openfief-trusted-signers` (D-45). The operator reads that registry over plain HTTP only because its address is a private IPv4 address, for which go-containerregistry allows HTTP; the nodes get the image through `ctr --plain-http` pulls, so containerd needs no registry configuration. No operator flag or setting was added for the tests.
    - *Options:* keep; or give the registry TLS with a test CA the operator and the nodes trust, closer to a real registry but with per-distribution node configuration.
