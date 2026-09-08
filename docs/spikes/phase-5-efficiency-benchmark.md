# Phase 5 efficiency benchmark

Date: 2026-09-08

This benchmark is an aggregate repository artifact, not canonical Cassor state.
`cassor benchmark --json` creates disposable SQLite fixtures, runs the same
quality gate for each scenario, and removes the fixtures before returning. It
does not persist transcripts, hidden reasoning, or run telemetry.

## Results

| Scenario | Outcome | Observed operations | Elapsed | Output bytes | Repair effort |
| --- | --- | ---: | ---: | ---: | ---: |
| Small local change | passed | 11 | 3.136 ms | 19 | 0 |
| New feature | passed | 7 | 2.328 ms | 1,053 | 0 |
| Approved plan amendment | passed | 4 | 1.552 ms | 611 | 0 |
| Fresh-session resumption | passed | 2 | 0.848 ms | 838 | 0 |

The elapsed values are one observed local run and are directional rather than
performance guarantees. The quality gate for every row was successful scenario
completion without an error. The scenarios exercise the CLI's delivery
contract through the store layer: approval, task lifecycle, evidence,
amendment, and bounded scoped context.

## Telemetry boundary

This Codex runtime did not expose safe worker isolation, a worker comparison
run, or token telemetry. Those dimensions are recorded as `unavailable`; no
values are estimated or treated as zero. The benchmark therefore establishes a
single-agent baseline and a repeatable scenario contract, not a claim that one
runtime or model is faster than another.

The benchmark command is intentionally non-persistent. If a future runtime
exposes comparable worker execution and token metrics, the same four scenarios
can be run from equivalent fixtures and compared first by successful outcome,
then by elapsed time, observed calls, repair effort, and reported tokens.
