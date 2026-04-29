# 🛰️ EventGuard CLI: The Governance Engine

The EventGuard CLI is the control plane for your organization's **Data-as-Code** ecosystem. It enables Analysts to maintain high-integrity tracking plans while providing Engineers with the binaries required for sub-millisecond validation.

---

## 🛠️ Core Capabilities

### 1. `init`: Bootstrap Governance
Initializes a new tracking repository with an immutable audit trail.
- Sets up a deterministic `canvas.yaml`.
- Configures local Git state for automated branch management.

### 2. `propose`: Analyst-First Workflows
The `propose` command abstracts the complexity of Git, allowing Data Analysts to contribute to the codebase safely.
- **Impact Checking**: Automatically runs a parity check against current production snapshots.
- **Branch Management**: Isolates changes into `analyst/` branches for CI/CD review.

### 3. `serve`: Zero-Downtime Proxy
A high-performance validation middleware that sits between your trackers and your data lake.
- **Live Reload**: Hot-swap tracking plans via `POST /admin/reload` without dropping events.
- **Pluggable Sinks**: Route invalid telemetry to S3, Kinesis, or Webhooks for debugging.

### 4. `export-wasm`: Browser-Side Shift-Left
Compiles the validation engine into a browser-compatible WASM binary.
- **`--hash`**: Uses SHA256 hashing to obfuscate event names and business logic, making it safe for public distribution.

---

## 🏗️ Technical Specifications (Event Canvas Standard)
The CLI is built for environments where data integrity is non-negotiable.

- **Language**: Go 1.26+
- **Throughput**: ~45k events/sec (Single Core)
- **Governance**: Integrated Integrity Guard (Identity Enforcement)

---

Built by the **Event Canvas Team**.
