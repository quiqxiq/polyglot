# ADR 0010: Protobuf Wire Compatibility

## Status
Accepted

## Decision
Existing protobuf field numbers, names, services, and wire meanings are immutable. New fields are additive and generated code is never edited manually. Breaking changes require explicit versioning and approval.

## Consequence
Older clients remain decodable. `buf lint` and generated-output diff checks stay mandatory.
