# ADR 0007: Pagination Contract

## Status
Accepted

## Decision
New or expanded list contracts use additive `page_size`, `page_token`, and `next_page_token` fields. Existing fields and semantics remain unchanged until an endpoint migration is agreed.

## Consequence
Large lists gain bounded reads without breaking current clients. Each list endpoint documents limits, ordering, and token ownership.
