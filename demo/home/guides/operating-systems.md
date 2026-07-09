# Operating Systems — Study Guide

## Summary

A tour of what an operating system does: run many programs on shared hardware
without letting them corrupt each other. Covers **processes and threads**,
**CPU scheduling**, and **virtual memory**.

## Key Concepts

- **Process vs. thread** — A process owns an address space; threads within it
  share that space but run independently.
- **Context switch** — Saving one process's registers and restoring another's;
  cheap between threads, costlier between processes.
- **Scheduling** — Round-robin (fair time slices) vs. priority (important jobs
  first, risking starvation).
- **Virtual memory** — Paging maps per-process virtual addresses to physical
  frames; a page fault pulls a page from disk.

## Common Traps

- **Confusing concurrency with parallelism** — concurrency is *structure*
  (interleaving), parallelism is *execution* (simultaneous).
- **Ignoring starvation** — pure priority scheduling can starve low-priority
  jobs forever without aging.
