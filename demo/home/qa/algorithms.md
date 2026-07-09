# Algorithms — Revision Q&A

## Q1. What does Big-O notation describe, and why do we drop constants?

**Big-O** describes an **upper bound** on how an algorithm's cost grows as the
input size *n* grows. We drop constant factors and lower-order terms because
asymptotic growth is what dominates for large *n* and it keeps the measure
independent of hardware and language.

## Q2. State the lower bound for comparison sorts and explain it in one line.

Any comparison sort needs **Ω(n log n)** comparisons in the worst case: there are
*n!* possible orderings, and a binary decision tree distinguishing them must have
height at least log₂(n!) = Θ(n log n).

## Q3. When does quicksort degrade to O(n²), and how do you avoid it?

Quicksort hits **O(n²)** when the pivot repeatedly splits off just one element —
classically on already-sorted input with a first/last pivot. Avoid it with a
**randomised** or **median-of-three** pivot.

## Q4. Compare merge sort and quicksort on space and stability.

**Merge sort** is stable and always O(n log n), but needs **O(n)** extra space.
**Quicksort** sorts **in place** (O(log n) stack) and is usually faster in
practice, but is **not stable** and has an O(n²) worst case.

## Q5. Which traversal finds the shortest path in an unweighted graph, and why?

**BFS**. It explores vertices in order of increasing distance from the source, so
the first time it reaches a vertex it has done so by a path with the fewest
edges. It runs in O(V + E).

---

Keep going — you know this cold. 🎯
