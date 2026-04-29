# EventGuard: How to Use

EventGuard provides specialized workflows for every member of the product and data team.

## 1. For Analysts (Taxonomy Design)

Tracking plans live in `maps/*.yaml`. Use **Mixins** to avoid duplication.

### Designing a Contract
```yaml
taxonomy:
  mixins:
    Identity:
      properties:
        userId: { type: string, required: true }
  events:
    Login:
      imports: [Identity]
      properties:
        method: { type: string, enum: ["email", "google"] }
```

### Checking for Breaking Changes
Run `impact-check` in your PR to verify your changes won't break existing telemetry.
```bash
event_guard impact-check --prev-sha main
```

---

## 2. For Frontend Developers (WASM Shift-Left)

Validate events in the browser using the **exact same** logic as the backend.

### Exporting for Web
```bash
event_guard export-wasm --hash
```
This generates `validator.wasm`. Load it into your SDK to catch errors before ingestion.

---

## 3. For Backend Developers (Go Middleware)

Integrate EventGuard into high-throughput ingestion pipes.

### Validation Loop
```go
import "github.com/randodev95/event_guard/pkg/validator"

// 1. Initialize engine (usually during startup)
engine := validator.NewEngine(plan)
engine.Warmup()

// 2. Validate in hot path
res, err := engine.ValidateJSON(payload)
if !res.Valid {
    // Handle invalid data (e.g., push to DLQ)
    log.Printf("Invalid event: %v", res.Errors)
}
```
**Performance**: Full validation takes **~11.8μs**.

---

## 4. For Product Teams (Visualization)

EventGuard automatically generates documentation and flow diagrams.

### Generate Docs
```bash
event_guard generate --target html --output docs.html
event_guard generate --target mermaid --output flows.mmd
```
These outputs provide a source-of-truth for what is being tracked and how users flow through the system.

---

## CLI Reference

- `init`: Setup a new project.
- `propose -m "msg"`: Auto-branch and commit tracking plan changes.
- `validate <file>`: Test a sample payload locally.
- `serve`: Start a validation microservice.
