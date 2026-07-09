# Algorithms — Study Guide

## Summary

This course builds the toolkit for **analysing** and **choosing** algorithms.
It opens with **complexity** (Ch.1): measuring how time and memory scale with
input size, and the asymptotic notation that keeps only the dominant term. It
then covers **sorting** (Ch.2): the comparison-sort lower bound and the
trade-offs between insertion, merge, quick, and heap sort. It closes with
**graphs** (Ch.3): representing networks and traversing them with BFS, DFS, and
Dijkstra. The throughline: reason about cost *before* reaching for code.

## Key Concepts

### Complexity (Chapter 1)

- **Asymptotic analysis** — Describe growth as *n → ∞*, dropping constants and
  lower-order terms, so the measure is hardware-independent.
- **Big-O / Ω / Θ** — Upper bound, lower bound, and tight bound on growth.
- **Growth hierarchy** — `O(1) < O(log n) < O(n) < O(n log n) < O(n²) < O(2ⁿ)`.

### Sorting (Chapter 2)

- **Comparison-sort bound** — No comparison sort beats **Ω(n log n)**.
- **Stability** — A stable sort preserves the relative order of equal keys.
- **In-place** — Uses O(1) extra space beyond the input.

### Graphs (Chapter 3)

- **Representation** — Adjacency list (sparse) vs. adjacency matrix (dense).
- **BFS vs. DFS** — Level-order (shortest unweighted path) vs. depth-first
  (topological sort, cycle detection). Both O(V + E).

## Formulas

| Algorithm       | Best        | Average       | Worst       | Space   |
|-----------------|-------------|---------------|-------------|---------|
| Insertion sort  | O(n)        | O(n²)         | O(n²)       | O(1)    |
| Merge sort      | O(n log n)  | O(n log n)    | O(n log n)  | O(n)    |
| Quicksort       | O(n log n)  | O(n log n)    | O(n²)       | O(log n)|
| Heapsort        | O(n log n)  | O(n log n)    | O(n log n)  | O(1)    |

## Common Traps

- **Dropping the wrong term** — keep the *fastest-growing* term, not the first.
- **"Quicksort is always fast"** — its worst case is O(n²); a bad pivot on
  already-sorted input triggers it.
- **Dijkstra with negative edges** — it silently gives wrong answers; use
  Bellman–Ford instead.
