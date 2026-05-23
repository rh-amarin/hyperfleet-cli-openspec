## ADDED Requirements

### Requirement: POST proxy for cluster adapter statuses
The server SHALL accept `POST /api/clusters/{id}/statuses` requests, read the JSON body, and forward it verbatim to the upstream HyperFleet API at `clusters/{id}/statuses`. The upstream response SHALL be forwarded to the browser with the same HTTP status code and `Content-Type: application/json`.

#### Scenario: Successful POST proxy
- **WHEN** browser sends POST /api/clusters/abc/statuses with a valid AdapterStatusCreateRequest body
- **THEN** server forwards the body to the upstream and returns the upstream 2xx response

#### Scenario: Upstream validation error forwarded
- **WHEN** upstream returns HTTP 422 for an invalid request
- **THEN** server returns HTTP 422 with the upstream JSON error body

### Requirement: POST proxy for nodepool adapter statuses
The server SHALL accept `POST /api/clusters/{id}/nodepools/{npid}/statuses` requests and proxy them to `clusters/{id}/nodepools/{npid}/statuses` with the same forwarding semantics as the cluster status POST.

#### Scenario: Successful nodepool POST proxy
- **WHEN** browser sends POST /api/clusters/abc/nodepools/np1/statuses with a valid body
- **THEN** server forwards to the correct upstream path and returns the upstream response

### Requirement: GET-only guard preserved for non-status routes
All existing routes that are not status-creation endpoints SHALL continue to reject non-GET methods with HTTP 405.

#### Scenario: POST to cluster list rejected
- **WHEN** browser sends POST /api/clusters
- **THEN** server returns HTTP 405
