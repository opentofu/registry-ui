# Search worker

The Search worker is used as the backend that route requests on our Cloudflare Worker. Static files are served from R2 directly and search goes to our serverless Neon Database. There's a cache in Cloudflare worker as well that is handled by this application.

## Installation

To setup your local environment for this project:

1. Copy the `.example.dev.vars` to a new `.dev.vars` file in the same folder.
2. When running docker-compose, the whole folder will be copied, so you can overwrite `.dev.vars` with your environment variables.

## Feeding data

There's a `Makefile` folder on the root with commands to setup the data in this application.
