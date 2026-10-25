# generation-fence-single-probe

Objective: remove the redundant pre-`Next` generation probe so each row pays one filesystem check instead of two.
Branch: `perf/generation-fence-single-probe`
Delivery: work-unit commit on the branch; push and PR remain a human decision.
Test runner: `go test ./internal/store`

## Specs

- **S1**: `generationRows.Next` performs exactly one generation probe per row, after the driver row read.
- **S2**: The documented invariant is unchanged: replacing the database file while iterating still fails the same `Next` call (`TestGenerationFenceRejectsUnsafeOperations/rejects_while_iterating_rows`).
- **S3**: Every other fence call site keeps its probes: `QueryContext`, `ExecContext`, `BeginTx`, `Commit`, `Rollback`, `Close`, `Ping`, `ResetSession`, and the `FileControl` methods.
- **S4**: The pre-probe is redundant because `QueryContext` probes after the query and the previous row's post-probe covers the window before the next row.
- **S5**: No behavior change other than the removed duplicate probe.

## Tasks

- [x] T1 measure the baseline benchmark and record it — 99.84s before, same machine
- [x] T2 remove the pre-`Next` probe
- [x] T3 focused invariant tests and benchmark comparison — fence tests PASS; 22.55s after (4.4x)
- [ ] T4 independent verification and commit

## Log

- **L1** (user request, verbatim): "si procede"
- **L2**: Audit measurement on this machine: `BenchmarkScanProject_Page5000` = 99.84s with both probes; 20.83s with only the pre-probe removed (4.8x); 1.98s with no per-row probes (50x, breaks S2). pprof attributed 86% of CPU to `databaseGeneration.check` and 64% to `generationRows.Next`.
