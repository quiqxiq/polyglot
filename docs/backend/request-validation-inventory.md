# Request Validation Inventory

Source of truth for request validation at the protobuf boundary. Wire format
does not change here. New constraints must be additive and validated with
`buf lint` plus generated-code diff checks.

## Rules

- `wire`: protovalidate constraint belongs in `api/proto/v1/*.proto`.
- `business`: rule needs domain state, repository, actor, or cross-field logic.
- `transport`: header, cookie, or authentication parsing remains in adapter.
- `covered`: constraint exists and is exercised by a contract test.
- `gap`: constraint or contract test is still missing.

## Current Inventory

| Context | Request family | Wire constraints | Business/transport rules | Status |
|---|---|---|---|---|
| auth | login, refresh, profile, password | username/password/refresh fields partially covered | bearer header and cookie parsing | gap |
| user | create, update, delete, device assignment | create/update fields and IDs partially covered | role and actor authorization | gap |
| customer | CRUD and lookup | IDs, lookup identifiers, nested customer partially covered | customer uniqueness and lifecycle | gap |
| device | get, update, delete, streams, metrics | IDs and nested device partially covered | driver access, metric time range | partial |
| billing | invoice, cashier, plan, subscription | IDs, amounts, required nested plan covered in many requests | payment state and subscription transitions | partial |
| hotspot | users, profiles, DHCP, vouchers, streams | device IDs and selected resource IDs partially covered | mutually exclusive filters and router rules | gap |
| PPP | secrets, profiles, sessions, streams | device IDs, RouterOS IDs, names partially covered | driver capability and RouterOS state | gap |
| notification | templates, queue, send | selected IDs and content fields partially covered | sender availability and retry state | gap |
| bot | conversations, WhatsApp, LLM, skills | selected IDs and nested values partially covered | session state, permissions, provider availability | gap |
| registration | submit, approval, install, conversion | selected IDs and fields partially covered | status transitions and provisioning | gap |
| cashbook | account, category, transaction | selected nested fields partially covered | account/category existence and amount rules | gap |
| reporting | daily, monthly, yearly, snapshot | date/year constraints incomplete | repository availability and snapshot state | gap |
| settings | category, key, batch, bot settings | category/key/repeated settings covered | setting ownership and value semantics | partial |

## Contract Tests

Existing boundary tests:

- `internal/adapter/connect/device/validation_contract_test.go` covers invalid
  unary `GetDevice` and invalid server-stream `StreamPing`.
- `pkg/response/http_test.go` covers HTTP fault mapping.
- `internal/adapter/http/adminapi/handler_test.go` covers missing query input.
- `internal/adapter/http/gateway/handler_test.go` covers malformed JSON.
- `internal/adapter/http/reports/handler_test.go` covers invalid date.

Required next tests:

- one invalid request per ConnectRPC bounded context;
- one invalid server-stream request per streaming service family;
- nested required message rejection;
- numeric range and repeated-item rejection;
- HTTP unknown-field and oversized-body rejection;
- proof that invalid wire requests do not invoke usecase dependencies.

## ID Policy Gaps

The following decisions remain open before adding broad regex constraints:

- whether internal IDs are UUIDs, numeric strings, or RouterOS `.id` values;
- maximum length for user-supplied identifiers;
- whitespace normalization policy;
- whether public phone identifiers accept local and international forms;
- whether legacy RPC message names can remain permanently for wire compatibility.
