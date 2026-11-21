# Agent instructions

## Testing policy

Before adding, reviewing, refactoring, consolidating, or deleting tests, or changing application behavior, read and follow [Testing Business Rules](.claude/skills/testing-business-rules/SKILL.md).

- Identify the rule, its evidence, and the failure consequence. Test observable behavior at the smallest boundary that honestly proves it.
- Prefer cohesive scenarios that share expensive setup and cover a related correction, retry, or save lifecycle. Do not combine independent risks just to reduce the test count.
- Remove tests only after establishing that they protect no valid contract or retained coverage detects the same failure. Do not declare edge cases impossible without evidence.
- Preserve distinct authorization, privacy, data-integrity, monetary, concurrency, persistence, provider-contract, and real regression guarantees.
- Run the affected suites and relevant integration tests. Report actual commands, results, and any gaps; when optimizing runtime, measure comparable runs rather than equating fewer tests with faster CI.

Nested AGENTS.md files provide additional instructions for their directories.