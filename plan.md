1. Refactor `api/internal/github/checks.go` to implement `CreatePendingChecks` and `UpdateChecks` methods for granular check suites, as outlined in the objective.
2. Refactor `api/internal/github/client.go` to add `conclusion string` to the arguments of `CreateCheckRun` and pass it to the JSON payload, along with fixing all references.
3. Write extensive tests in `api/internal/github/checks_test.go` to mock the GitHub API and ensure that the right POST / PATCH requests are sent, and advisory logic works correctly based on `substrate.yaml` configs.
4. Ensure code passes tests using `go test ./...`.
5. Pre-commit check to ensure `go vet` and format pass.
6. Submit the PR.
