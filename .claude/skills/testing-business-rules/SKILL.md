---
name: testing-business-rules
description: >
  Use when adding, reviewing, or deleting tests, deciding whether a function needs coverage, or when a suite enumerates branches, signs, empty values, or implementation steps without naming the behavior at risk.
---

# Testing Business Rules

A test earns its place by protecting a meaningful behavior or contract, not by demonstrating today's implementation. A rule can be a product decision, a technical guarantee, or a consumer expectation.

## Quick reference

| Keep coverage for | Reject or reshape |
| --- | --- |
| Domain behavior, API/provider contracts, numeric or ordering guarantees | Assertions that merely restate the implementation |
| Validation, authorization, privacy, data integrity | Input permutations with no distinct risk |
| Idempotency, retries, discard policies, persistence and queue integration | Private helper calls or incidental call order |
| Observed regressions, including obvious mistakes | Duplicate coverage with no additional failure detection |
| Boundaries with a named consequence | Snapshots that freeze incidental output |

Currency rounding can be a contract even when it uses simple arithmetic. Coverage metrics reveal gaps; they do not justify tests.

## Procedure

### 0. Scope suite-wide cleanup by module

For application-wide pruning, inventory the suites and actual CI commands first, then deliver independently verified module-sized changes rather than a sweeping deletion. Read a recent successful CI run's job steps and package timings; distinguish test execution from compilation, database/framework startup, browser installation, and rendering. Prioritize redundant expensive setup and rendering over deleting cheap parameter rows, because fewer tests do not necessarily shorten the critical path. Report completed modules separately from the remaining audit backlog.

When asked to distribute this policy across repositories, inspect each repository's existing skill location and root/nested `AGENTS.md` files. Install or update the current skill and add a concise root directive requiring agents to read and follow its relative path for test changes. Preserve nested instructions, verify that references resolve and copied content matches the intended source, and follow the isolated-change/PR workflow for each repository. Do not silently add unrelated testing mandates or permanent policy-scanner tests.

### 1. Name the rule before writing the test

Record these in the test plan or review, not as boilerplate comments:

```text
Rule: an empty tenant ID must never grant access to another tenant's records.
Established in: tenant-isolation security contract.
Failure: a caller reads another tenant's private records.
```

Find evidence in requirements, API documentation, callers, history, or incidents, not current output alone. Missing evidence requires investigation with owners and consumers, not coverage removal. Temporary characterization tests can support legacy refactoring; label unconfirmed assumptions instead of declaring permanent contracts.

### 2. Find the coverage gap

Identify the wrong behavior this test would catch. Compare failure modes, not lines executed: happy-path coverage does not cover an authorization bypass.

Keep regressions that protect valid rules and add detection, including obvious mistakes, manually detectable bugs, and bad decisions.

Trace each negative case through earlier guards to confirm that it reaches the boundary it claims to test. A generic error can come from request-size validation before provider-output parsing; inject the malformed output at the provider seam or move the assertion to the parser boundary rather than counting the earlier rejection as equivalent coverage. Require evidence that an edge case is unreachable before removing it as impossible.

### 3. Choose the smallest honest boundary

Use a public function or module for local rules. Use focused integration tests for collaboration, serialization, persistence, or queue guarantees. Use route/full-flow tests for behavior smaller tests cannot establish.

Prefer fast, deterministic real dependencies. Reuse a public deferred/controlled-promise implementation from an existing dependency when available; generic async test utilities do not belong in feature-specific fixtures. Check runtime and configured TypeScript libraries before selecting a newer built-in API. Use a maintained fake when real dependencies are costly or unavailable; use stubs or mocks for controlled responses or contract-relevant interactions. Double at the existing dependency seam. For UI integration, keep routing, providers, controls, and the generated API client real; inject controlled responses at the public transport boundary. Do not relocate router/UI mocks into setup helpers and call the result an integration test. Internal versus external is not the deciding factor: never replace the behavior or collaboration this test claims to verify.

A mocked database cannot prove transaction semantics. A mocked billing service cannot prove duplicate-charge prevention. Exercise the relevant production collaboration and verify doubles against real contracts separately. Keep tests isolated from live services; control clocks, randomness, and asynchronous completion instead of sleeping.

### 4. Write one focused behavior per test

Name the condition and promised outcome:

- Weak: `handles empty input`
- Better: `denies cross-tenant access when the tenant ID is empty`

Separate setup, user actions, intermediate outcomes, retry actions, and final assertions with blank lines. Keep statements in the same logical phase together; use whitespace rather than explanatory phase comments for this user's frontend tests.

Use multiple assertions when they jointly establish one outcome. Use named parameterized cases for meaningful variants of one rule. Choose distinct, non-default values to expose ignored or swapped inputs; include zero, negative, empty, and boundary values when they challenge the contract.

When outcomes exclude each other, assert that the other outcome is absent. A test that checks only that the item expired still passes when the code also fulfills it. When a change adds an outcome, add its absence check to each existing test that expects no outcome. When a change moves a case from one outcome to another, name both in the test: `expires, not fulfills, once the re-read cap is reached`.

Assert observable results, state, or contractual interactions. Derive expected values independently of production logic. Avoid private structure and call order unless order itself is the guarantee. Names should explain arrange/act/assert; comments explain only non-obvious reasons.

### 5. Verify, then prune

Observe new tests fail for the intended missing behavior, then pass with the implementation. Check that existing tests would reject a plausible rule-breaking change. For changes to a shared lifecycle or completion primitive, enumerate its production callers and test the distinct lifetime boundaries: same-session navigation, identity replacement, partial success and a newer concurrent write. A passing example at one caller does not establish root-cause closure across the application. Run focused tests and relevant integration tests; report gaps.

For migrations repairing concurrency tokens or immutable lifecycle records, seed populated legacy histories and execute the actual migration with database guards enabled. Assert both the intended token change and the absence of content/status/snapshot/artifact changes; fresh-database startup alone cannot verify upgrade safety.

Do not add repository-wide text scanners or source-file assertions for a one-off branding/copy replacement unless an explicit ongoing contract requires them. Verify the replacement during the task instead; permanent copy policing adds maintenance without protecting business behavior.

Never delete by title: rename a poorly named security test. Remove a test only if it protects no valid contract or another test detects the same failure mode. Verify that conclusion first. Preserve distinct integration risks even when local tests cover the rule.

When consolidating expensive route tests, move rejection assertions into the corresponding correction/retry lifecycle rather than dropping them. Preserve both same-tick and later pending submissions, request totals after completion, and the next intentional save. Compare original and consolidated tests against temporary rule-breaking mutations, restore production code, and measure alternating warm runs with identical dependencies/configuration; report test-body and wall time separately without extrapolating local savings to all CI.

## Report

Summarize by rule, not every parameterized row:

```text
Protects: <rule; evidence; failure consequence>
Boundary: <function | module | integration | route/task; why; doubles>
Changed: <added/kept/renamed/removed; equivalent coverage for removals>
Verified: <commands and results; gaps or unconfirmed assumptions>
```

## Google guidance

This is a local policy informed by Google's [Testing on the Toilet](https://testing.googleblog.com/search/label/TotT), not a formal Google standard:

- [Test Behavior, Not Implementation](https://testing.googleblog.com/2013/08/testing-on-toilet-test-behavior-not.html): public contracts survive implementation changes, including simple arithmetic APIs.
- [Test Behaviors, Not Methods](https://testing.googleblog.com/2014/04/testing-on-toilet-test-behaviors-not.html): separate behaviors, not one test per method or assertion.
- [Don't Overuse Mocks](https://testing.googleblog.com/2013/05/testing-on-toilet-dont-overuse-mocks.html): prefer real dependencies or fakes when practical; mocks can drift and obscure intent.
- [Just Say No to More End-to-End Tests](https://testing.googleblog.com/2015/04/just-say-no-to-more-end-to-end-tests.html): related Google Testing Blog guidance on fast, reliable unit tests and focused integration tests.
- [Choosing Values for Robust Tests](https://testing.googleblog.com/2026/06/choosing-values-for-robust-tests.html): distinct, non-default inputs and meaningful boundary cases expose false positives.
