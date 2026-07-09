# Chapter 2 — Sorting

Comparison sorts order elements using pairwise comparisons; the lower bound for
any comparison sort is Ω(n log n).

- **Insertion sort** — O(n²), but O(n) on nearly-sorted input; stable, in-place.
- **Merge sort** — O(n log n) always; stable; needs O(n) extra space.
- **Quicksort** — O(n log n) average, O(n²) worst; in-place; pivot choice matters.
- **Heapsort** — O(n log n) always; in-place; not stable.

Non-comparison sorts (counting, radix) beat the bound when keys are bounded.
