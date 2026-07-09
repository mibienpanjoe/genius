# Chapter 3 — Graphs

A graph G = (V, E) is a set of vertices and edges. Traversals visit every
reachable vertex once.

- **BFS** — queue-based, explores level by level, finds shortest paths in
  unweighted graphs. O(V + E).
- **DFS** — stack/recursion, goes deep first; used for topological sort and
  cycle detection. O(V + E).
- **Dijkstra** — shortest paths with non-negative weights, O((V + E) log V) with
  a binary heap.
