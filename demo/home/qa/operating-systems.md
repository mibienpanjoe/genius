# Operating Systems — Revision Q&A

## Q1. What is the difference between a process and a thread?

A **process** is a program in execution with its own private address space. A
**thread** is a unit of execution *inside* a process; threads share the process's
address space but have their own stack and registers.

## Q2. What problem does virtual memory solve?

**Virtual memory** gives each process the illusion of a large, contiguous, private
address space. Paging maps virtual addresses to physical frames and lets the sum
of all processes' memory exceed physical RAM by backing pages on disk.

---

Solid. On to the next topic. 🎯
