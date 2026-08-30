# ADR 0009: Request Validation Boundary

## Status
Accepted

## Decision
Protovalidate handles wire constraints at ConnectRPC boundary. Usecases handle business rules requiring state, repositories, or actor context. HTTP handlers validate DTO shape and decoder limits, not duplicated business rules.

## Consequence
Invalid wire requests fail before usecase execution. Business validation remains transport-independent and reusable.
