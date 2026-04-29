# EventGuard: Implementation Guide

This guide provides a problem-oriented approach to integrating EventGuard into your data pipeline.

## 1. Local Development (For Developers & Analysts)

The EventGuard CLI is your primary tool for managing tracking plans and validating events locally.

### Initialize a Project
Set up a standard directory structure and genesis tracking plan.
```bash
event_guard init
```

### Validate a Sample Event
Verify that a local JSON payload conforms to your tracking plan.
```bash
event_guard validate payload.json --plan maps/
```

### Propose a Change
Automate the creation of a new Git branch and commit for your tracking plan updates.
```bash
event_guard propose -m "Add discount_code to Purchase event"
```

---

## 2. CI/CD Integration (For Data Engineers)

Prevent breaking changes from reaching production by enforcing the tracking contract in your CI pipeline.

### Impact Check
Compare the current plan against previous snapshots to detect breaking schema changes.
```bash
event_guard impact-check --prev-sha main
```
*Tip: Integrate this into your GitHub Actions to automatically block PRs that violate the contract.*

---

## 3. Production Deployment (For Platform Engineers)

Enforce validation at scale using the high-performance Go engine or the validation microservice.

### Running the Microservice
Start a low-latency validation server with live-reload support.
```bash
event_guard serve --port 8080 --plan maps/
```

### In-Process Validation (Go)
For maximum performance, use the validator directly in your ingestion pipeline.
```go
engine := validator.NewEngine(plan)
res, err := engine.ValidateJSON(payload)
```
*Note: In-process validation typically averages **11.8μs**.*

---

## 4. Frontend & Browser (For Web Developers)

Shift validation to the client-side to catch errors before they are sent over the wire.

### Exporting to WASM
Compile your tracking plan and validation logic into a single WebAssembly binary.
```bash
event_guard export-wasm --hash
```
The `--hash` flag obfuscates event names for security, ensuring your business logic isn't exposed in the client-side bundle.

---

## 5. Documentation & Visualization (For Product Teams)

Keep your documentation in sync with your implementation automatically.

### Generate Visualizations
```bash
# Generate HTML Documentation
event_guard generate --target html --output docs/index.html

# Generate Mermaid Flow Diagrams
event_guard generate --target mermaid --output docs/flows.mmd
```
