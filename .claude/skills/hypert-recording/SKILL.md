---
name: hypert-recording
description: Record or re-record HTTP fixtures for OpenFief integration tests with hypert, safely, with no real account data. Use when adding an HTTP adapter test or when a provider's API changed.
---

# Recording HTTP fixtures with hypert

Integration tests replay committed recordings offline (docs/dev/testing.md). Recording hits the real API, so it's a deliberate, manual act.

1. **Pick a safe source.** Use public endpoints (NBP, GUS), a vendor sandbox, or an IBKR paper account. Never a real account with real holdings. If only a real account can produce the response, stop and ask the owner.
2. **Configure sanitizing before the first run:**
   - **Requests:** use `hypert.WithRequestSanitizer(hypert.ComposedRequestSanitizer(hypert.DefaultRequestSanitizer(), hypert.HeadersSanitizer(...), hypert.SanitizerQueryParams(...)))` for every token, key or account parameter the API uses.
   - **Responses:** use `hypert.WithResponseTransform(...)` in record mode to replace account numbers, names and IDs. hypert does not sanitize responses on its own.
3. **Record:** flip the test's record flag locally (`hypert.TestClient(t, true, ...)`), run that one test, then flip it back to `false`. Never commit `true`.
4. **Review every file** in `testdata/<TestName>/` before staging. `git grep` the recordings for the token, the account number and the owner's name. All must be absent.
5. **Replay offline:** run the test again with the network off. It must pass and validate each request.
6. **Commit** the recordings with the test, in the same change.
