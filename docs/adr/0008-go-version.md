# ADR 0008: Supported Go Version

## Status
Accepted

## Decision
Project targets Go `1.26`. `go.mod`, CI, Makefile guidance, and developer documentation must use this version.

## Consequence
New language and standard-library features are allowed only within Go 1.26. Version drift is a CI defect.
