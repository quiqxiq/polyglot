// Package tenant defines tenant identity for this single-tenant deployment.
//
// Polyglot runs as a single-tenant application. The DefaultID value is kept
// for backwards compatibility with legacy database columns and protobuf
// fields; new code should not introduce additional tenant identifiers.
package tenant

// DefaultID is the canonical tenant identifier used by the single-tenant
// deployment. Legacy rows, protobuf fields, and database defaults still carry
// this value for wire and schema compatibility.
const DefaultID = "tenant-default"
