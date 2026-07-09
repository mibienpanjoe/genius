# Chapter 1 — Complexity & Big-O

Algorithm analysis measures how running time and memory grow with input size *n*,
independent of hardware. We use asymptotic notation to keep only the dominant term.

- **Big-O** — upper bound on growth (worst case).
- **Big-Ω** — lower bound.
- **Big-Θ** — tight bound (matching upper and lower).

Common classes, slowest-growing first: O(1) < O(log n) < O(n) < O(n log n) <
O(n²) < O(2ⁿ) < O(n!). Constant factors and lower-order terms are dropped.
