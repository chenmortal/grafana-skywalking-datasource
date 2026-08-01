# Skywalking Datasource for Grafana

English | [简体中文](README_zh-Hans.md)

![Dynamic JSON Badge](https://img.shields.io/badge/dynamic/json?logo=grafana&query=%24.version&url=https%3A%2F%2Fgrafana.com%2Fapi%2Fplugins%2Fchenmortal-skywalking-datasource&label=Marketplace&prefix=v&color=F47A20)
![Dynamic JSON Badge](https://img.shields.io/badge/dynamic/json?logo=grafana&query=%24.grafanaDependency&url=https%3A%2F%2Fgrafana.com%2Fapi%2Fplugins%2Fchenmortal-skywalking-datasource&label=Grafana&color=F47A20)

A Grafana data source plugin that connects to an **Apache SkyWalking** OAP server via GraphQL API, enabling distributed tracing visualization directly in Grafana.

## Overview

This plugin brings SkyWalking's distributed tracing data into Grafana, allowing you to:

- **Search traces** across layers, services, endpoints, and instances
- **Visualize individual traces** with Grafana's native trace view
- **Filter traces** by trace state (All / Success / Error) and sort by duration or start time
- **Navigate from search results to detailed traces** via built-in data links

The plugin supports both SkyWalking v1 (`queryBasicTraces` / `queryTrace`) and v2 (`queryTraces`) GraphQL APIs, configurable via a toggle.

## Requirements

- **Grafana** `>= 12.1.0`
- **Apache SkyWalking OAP** server with GraphQL endpoint accessible from Grafana

## Getting Started

### Installation

1. Install the plugin from the [Grafana Plugin Catalog](https://grafana.com/grafana/plugins/chenmortal-skywalking-datasource/) or by downloading the release archive.
2. Place the plugin in your Grafana plugins directory (or follow the [Grafana plugin installation guide](https://grafana.com/docs/grafana/latest/administration/plugin-management/)).

### Configuration

1. In Grafana, navigate to **Administration** → **Plugins and data** → **Data sources** → **Add data source**.
2. Search for **Skywalking** and select it.
3. Fill in the configuration:

| Setting                  | Description                                  | Example                                     |
| ------------------------ | -------------------------------------------- | ------------------------------------------- |
| **URL**                  | SkyWalking OAP GraphQL endpoint              | `http://skywalking-oap.example.com/graphql` |
| **Path**                 | Additional resource path                     | `/resources` (default)                      |
| **Interface Version v2** | Enable to use SkyWalking v2 Query Traces API | `true` / `false`                            |
| **API Key**              | Authentication token (optional)              | `your-api-key`                              |

> **Note:** Once the Interface Version v2 toggle is enabled, it cannot be reverted. This setting determines which SkyWalking API is used for trace queries.

### Provisioning

You can provision the datasource via Grafana's provisioning system:

```yaml
datasources:
  - name: 'skywalking'
    type: 'chenmortal-skywalking-datasource'
    access: proxy
    url: 'http://skywalking-oap:12800/graphql'
    jsonData:
      path: '/resources'
      interfacev2: false
    secureJsonData:
      apiKey: 'your-api-key'
```

## Query Types

The plugin supports two query modes:

### Search

Search traces across your SkyWalking infrastructure with the following filters:

- **Layer** — Select the layer (e.g., GENERAL, VIRTUAL_MQ, etc.)
- **Service** — Filter by service name
- **Endpoint** — Search endpoints by keyword
- **Instance** — Filter by specific service instance
- **Trace State** — All / Success / Error
- **Query Order** — Sort by duration (slowest first) or start time (newest first)

### Trace ID

Look up a specific trace by entering its Trace ID. The trace will be displayed using Grafana's native trace visualization with full span details including service tags, span tags, references, and timing information.

## Dashboard Integration

The trace search results include links to individual trace views. When you select a trace from the search results, Grafana opens a detailed trace view showing:

- Span tree with parent-child relationships
- Operation names, service names, and instance information
- Duration breakdown for each span
- Span tags (peer, layer, component, and custom tags)
- Reference relationships (CHILD_OF, FOLLOWS_FROM)

## Development

### Prerequisites

- **Go** `>= 1.26`
- **Node.js** `>= 22`
- **npm** `>= 11`
- **Mage** (Go build tool)

### Setup

```bash
# Install frontend dependencies
npm install

# Install Go dependencies
go mod download

# Generate GraphQL types (backend)
go generate ./pkg/...

# Generate GraphQL types (frontend)
npm run codegen
```

### Build

```bash
# Build both frontend and backend
npm run build

# Development mode with watch
npm run dev
```

### Testing

```bash
# Run frontend unit tests
npm test

# Run frontend tests for CI
npm run test:ci

# Run Go backend tests
go test ./...

# Run end-to-end tests
npm run e2e
```

### Lint & Type Check

```bash
# TypeScript type checking
npm run typecheck

# ESLint
npm run lint

# ESLint with auto-fix
npm run lint:fix
```

### Local Development with Docker

```bash
# Start Grafana with the plugin mounted
npm run server
```

This starts a Grafana instance via Docker Compose with the plugin automatically loaded.

### Project Structure

```
├── pkg/                          # Go backend plugin
│   ├── main.go                   # Entry point
│   └── plugin/
│       ├── datasource.go         # Datasource lifecycle
│       ├── skywalking.go         # Service layer, health checks
│       ├── client.go             # GraphQL API client
│       ├── query.go              # Query handling & trace frame transformation
│       ├── callresource.go       # HTTP resource routes
│       └── graphql.go            # Generated GraphQL types
├── src/                          # Frontend TypeScript/React
│   ├── module.ts                 # Plugin registration
│   ├── datasource.ts             # Data source class
│   ├── components/
│   │   ├── ConfigEditor.tsx      # Datasource configuration UI
│   │   ├── QueryEditor.tsx       # Query builder
│   │   └── SearchForm.tsx        # Trace search form
│   └── locales/                  # i18n translations (en-US, zh-Hans)
├── graphql/                      # GraphQL schema & query definitions
├── provisioning/                 # Grafana provisioning examples
└── tests/                        # Playwright e2e tests
```

The backend is written in Go and communicates with the SkyWalking OAP server via GraphQL. The frontend is a TypeScript/React application that provides the query editor UI and integrates with Grafana's trace visualization.

### GraphQL Code Generation

The project uses code generation for both backend and frontend:

- **Backend:** [genqlient](https://github.com/Khan/genqlient) generates Go types from GraphQL operations defined in `graphql/`
- **Frontend:** [@graphql-codegen](https://the-guild.dev/graphql/codegen) generates TypeScript types from the same GraphQL schema

To regenerate types after schema changes:

```bash
# Backend
go generate ./pkg/...

# Frontend
npm run codegen
```

## Contributing

Issues and pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the [Apache License 2.0](LICENSE).
