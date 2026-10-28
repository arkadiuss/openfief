---
name: hypert-recording
description: Record or re-record HTTP fixtures for OpenFief integration tests with hypert, safely, with no real account data. Use when adding an HTTP adapter test or when a provider's API changed.
---

# Recording HTTP fixtures with hypert

Integration tests replay committed recordings offline (docs/dev/testing.md). Recording hits the real API, so it is a deliberate, manual act.

1. **Pick a safe source:** public endpoints (NBP, GUS), a vendor sandbox, or an IBKR paper account. Never a real account with real holdings. If only a real account can produce the response, stop and ask the owner.
2. **Set up sanitizing before the first run.** Sanitize every token, key and account parameter in requests. hypert does not sanitize responses, so transform them in record mode to replace account numbers, names and IDs.
3. **Record** with the test's record switch turned on locally, run only that test, then turn it off again. Never commit it on.
4. **Review every recorded file before staging:** the token, the account number and the owner's name must all be absent.
5. **Replay offline:** run the test again without network access. It must pass and validate each request.
6. **Commit** the recordings with the test, in the same change.
