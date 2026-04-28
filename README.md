# EventCanvas 🌌

**Deterministic Tracking Plan Engine for High-Integrity Data Pipelines**

EventCanvas is a sophisticated AST-based engine designed to enforce formal business constraints on telemetry data. It bridges the gap between product requirements and data engineering by transforming a YAML-based Tracking Plan into executable warehouse configurations and living documentation.

[![Go Report Card](https://goreportcard.com/badge/github.com/randodev95/eventcanvas)](https://goreportcard.com/report/github.com/randodev95/eventcanvas)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

## ✨ Features

- **Formal DSL**: Define contexts, events, and complex user journeys (flows) in a high-level YAML format.
- **Strict Validation**: Enforce `NoNull`, `Unique`, `Range`, and `Enum` constraints at the schema level.
- **Recursive Inheritance**: Build modular contexts that events can inherit, reducing duplication and ensuring consistency.
- **Multi-Target Generation**:
  - 🛠 **DBT/SQLMesh**: Generate data warehouse models with built-in validation.
  - 📄 **HTML Docs**: Beautiful, searchable documentation for product and engineering teams.
  - 📊 **Mermaid Diagrams**: Visualize complex stateful journeys and flows.
- **Identity Integrity**: Prevents "Ghost Users" by enforcing mandatory identity properties across all events.

## 🚀 Quick Start

### 1. Define your plan (`canvas.yaml`)

```yaml
identity_properties: ["wallet_address"]

contexts:
  Wallet_Context:
    properties:
      wallet_address: { type: string, required: true, no_null: true }
      chain_id: { type: integer, required: true }

events:
  "Transaction Initiated":
    category: "TRANSACTION"
    inherits: ["Wallet_Context"]
    properties:
      tx_type: { type: string, enum: ["stake", "unstake"] }
```

### 2. Generate Documentation

```bash
eventcanvas generate -t html -p canvas.yaml -o docs.html
```

### 3. Generate Warehouse Models

```bash
eventcanvas generate -t dbt -p canvas.yaml -o models/staging/
```

## 🏗 Architecture

EventCanvas is built with a decoupled architecture to ensure scalability and maintainability:

- **`pkg/ast`**: The core logical representation of the tracking plan. Handles property resolution and integrity checks.
- **`pkg/parser`**: Handles the ingestion of YAML definitions into the AST.
- **`internal/generator`**: Pluggable generators for various downstream targets.
- **`internal/tui`**: Interactive dashboard for real-time plan exploration.

## 🛠 Development

### Prerequisites
- Go 1.21+

### Running Tests
```bash
go test ./...
```

### Building the CLI
```bash
go build -o eventcanvas main.go
```

## 🗺 Roadmap

- [ ] **Runtime Validator**: Go/TypeScript SDKs for real-time event validation.
- [ ] **Schema Registry Integration**: Push generated schemas to Confluent/AWS Glue.
- [ ] **Drift Detection**: Compare actual warehouse data against the tracking plan.
- [x] **Versioned Plans**: Semantic versioning for tracking plan evolutions.

## 📦 Use as a Go Library

EventCanvas is designed to be imported into your own Go backend services for real-time validation:

```bash
go get github.com/randodev95/eventcanvas
```

```go
import (
    "github.com/randodev95/eventcanvas/pkg/validator"
    "github.com/randodev95/eventcanvas/pkg/parser"
)

func main() {
    // Load your plan
    plan, _ := parser.ParseYAML(data)

    // Create a high-performance validation engine
    engine := validator.NewEngine(plan)

    // Validate incoming JSON
    result, err := engine.ValidateJSON(payload)
}
```

---

Built with ❤️ by the EventCanvas Team.
