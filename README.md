# OpenTofu Registry
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopentofu%2Fregistry-ui.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopentofu%2Fregistry-ui?ref=badge_shield)


This repository contains the code that powers the [OpenTofu Registry](https://registry.opentofu.org) and its API at [api.opentofu.org](https://api.opentofu.org).

## Architecture Overview

The registry system consists of four main components that work together to scrape, process, index, and serve OpenTofu modules and providers with their documentation.

## Components

### Backend (Go)
The content processing engine that scrapes documentation and schemas from provider/module repositories. It extracts documentation from multiple formats (legacy `website/docs/` and modern `docs/` directories), validates repository licenses for redistributability, and uses a custom OpenTofu binary to extract module schemas including variables, outputs, resources, and submodules. 

The backend also serves the OpenAPI specification, which is **manually maintained** at `backend/internal/server/openapi.yml` and **must be updated whenever new API endpoints are added**.

### Frontend (React/TypeScript) 
The web UI for registry.opentofu.org. Uses React/TypeScript with Vite for the build system. Lets users browse modules and providers, read documentation, and search through the registry content.

### Search/pg-indexer (Go)
The indexing service that takes processed registry data and dumps it into a Neon PostgreSQL database for search functionality. It reads the latest registry information from the backend processing pipeline and structures it for efficient querying by the search API.

### Search Worker (Cloudflare Worker)
Runs the entire api.opentofu.org backend (the name is misleading). Handles search requests by hitting the Postgres database, and serves everything else from R2 storage. Caches search results for 5 minutes and static files for an hour.

## Development

See the `documentation/` folder for detailed technical documentation on each component.

## Contributing

Please see the [contribution guide](CONTRIBUTING.md) for details on how to work with this codebase.


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopentofu%2Fregistry-ui.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopentofu%2Fregistry-ui?ref=badge_large)