# Algorithms · chapter2-sorting — Revision Q&A

## Q1. What makes a sorting algorithm "stable"?

A sort is **stable** if it preserves the relative order of elements that compare
equal. This matters when sorting by a secondary key after a primary one.

## Q2. Why is heapsort in-place but not stable?

Heapsort sorts using only the input array plus O(1) variables, so it is
**in-place**. Building and popping the heap moves equal keys past one another, so
it is **not stable**.

---

Nice focus on sorting. 🎯
