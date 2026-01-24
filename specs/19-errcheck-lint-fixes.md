# Specification: Errcheck Lint Fixes

## Jobs to Be Done
- All `errcheck` violations reported by `golangci-lint` in CI are resolved
- Production code properly handles or explicitly discards error return values
- Test code checks error returns from encoding/decoding and write operations to avoid silent failures

## Problem 1: Unchecked `v.BindEnv` in Config Loader

`internal/config/config.go#L89-90` — `v.BindEnv("url")` and `v.BindEnv("token")` return errors that are silently discarded:

```go
v.SetEnvPrefix("LINKDING")
v.BindEnv("url")
v.BindEnv("token")
```

While `BindEnv` is unlikely to fail in practice (it only errors if called with zero arguments), discarding the error violates `errcheck` and sets a bad precedent for production code.

### Implementation

Wrap both calls and return an error if either fails:

```go
v.SetEnvPrefix("LINKDING")
if err := v.BindEnv("url"); err != nil {
    return nil, fmt.Errorf("failed to bind env for url: %w", err)
}
if err := v.BindEnv("token"); err != nil {
    return nil, fmt.Errorf("failed to bind env for token: %w", err)
}
```

Affected files:
- `internal/config/config.go` — `Load` function

## Problem 2: Unchecked `w.Write` in API Client Test

`internal/api/client_test.go#L351` — the mock HTTP handler in `TestCreateTag_Duplicate` calls `w.Write([]byte(...))` without checking the error:

```go
w.WriteHeader(http.StatusBadRequest)
w.Write([]byte(`{"name":["Tag with this name already exists"]}`))
```

If `Write` fails, the test would pass for the wrong reason (client receives an empty body instead of the expected error JSON).

### Implementation

Check the error and fail the test if the write fails. Since this is inside an `http.HandlerFunc`, use a helper or simply assign and check:

```go
w.WriteHeader(http.StatusBadRequest)
if _, err := w.Write([]byte(`{"name":["Tag with this name already exists"]}`)); err != nil {
    t.Errorf("failed to write response: %v", err)
    return
}
```

Note: The handler closure can reference `t` from the outer test function.

Affected files:
- `internal/api/client_test.go` — `TestCreateTag_Duplicate`

## Problem 3: Unchecked `json.NewDecoder().Decode()` in Import Tests

`internal/export/import_test.go#L91, L259, L308` — mock server handlers decode request bodies without checking the error:

```go
json.NewDecoder(r.Body).Decode(&bookmark)
```

If decoding fails (malformed JSON, wrong struct), the test silently operates on a zero-value struct, potentially masking bugs in the import logic.

### Implementation

Check the decode error and fail the test:

```go
if err := json.NewDecoder(r.Body).Decode(&bookmark); err != nil {
    t.Errorf("failed to decode request body: %v", err)
    http.Error(w, "bad request", http.StatusBadRequest)
    return
}
```

Affected files:
- `internal/export/import_test.go` — mock handlers at lines 91, 259, 308

## Problem 4: Unchecked `json.NewEncoder().Encode()` in Export Tests

`internal/export/html_test.go#L38` and `internal/export/csv_test.go#L236, L316` — mock server handlers encode responses without checking the error:

```go
json.NewEncoder(w).Encode(response)
```

If encoding fails, the test client receives a truncated or empty response, potentially hiding bugs in the export parsing logic.

### Implementation

Check the encode error and fail the test:

```go
if err := json.NewEncoder(w).Encode(response); err != nil {
    t.Errorf("failed to encode response: %v", err)
    return
}
```

Affected files:
- `internal/export/html_test.go` — mock handler at line 38
- `internal/export/csv_test.go` — mock handlers at lines 236, 316

## Problem 5: Unchecked `csvWriter.Write` in CSV Test

`internal/export/csv_test.go#L165, L179` — a test helper writes CSV rows without checking errors:

```go
csvWriter.Write(header)
// ...
csvWriter.Write(row)
```

If `Write` fails (e.g., buffer error), the test operates on incomplete CSV data.

### Implementation

Check the error and fail the test:

```go
if err := csvWriter.Write(header); err != nil {
    t.Fatalf("failed to write CSV header: %v", err)
}
// ...
if err := csvWriter.Write(row); err != nil {
    t.Fatalf("failed to write CSV row: %v", err)
}
```

Affected files:
- `internal/export/csv_test.go` — test helper around lines 165, 179

## Success Criteria
- [ ] `internal/config/config.go` checks return values of both `v.BindEnv` calls
- [ ] `internal/api/client_test.go` checks the `w.Write` return in `TestCreateTag_Duplicate`
- [ ] `internal/export/import_test.go` checks all three `json.NewDecoder().Decode()` calls (lines 91, 259, 308)
- [ ] `internal/export/html_test.go` checks the `json.NewEncoder().Encode()` call (line 38)
- [ ] `internal/export/csv_test.go` checks both `json.NewEncoder().Encode()` calls (lines 236, 316)
- [ ] `internal/export/csv_test.go` checks both `csvWriter.Write` calls (lines 165, 179)
- [ ] `golangci-lint run ./...` passes with no `errcheck` violations
- [ ] All existing tests continue to pass (`go test ./...`)
