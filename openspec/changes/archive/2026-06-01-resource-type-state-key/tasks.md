## 1. Config model

- [x] 1.1 Set `StateKey = name` at parse time; stop requiring `state-key` in YAML
- [x] 1.2 Derive path placeholders from entity name; honor optional `path-param`
- [x] 1.3 Update `ClusterID()` / `NodePoolID()` to use `clusters` / `nodepools`
- [x] 1.4 Remove `state-key` from config template

## 2. Command updates

- [x] 2.1 Update pubsub, rabbitmq, logs, cluster.go, nodepool.go state key references
- [x] 2.2 Update `hf rs types` output to show entity name as state key

## 3. Tests and docs

- [x] 3.1 Update `resource_types_test.go` and related cmd tests
- [x] 3.2 Update README resource-types documentation

## 4. Verification

- [x] 4.1 `go test ./...` and `go vet ./...`
- [x] 4.2 Save verification proof; live check `hf rs types` and state key display
