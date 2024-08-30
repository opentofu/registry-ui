# OpenTofu registry user interface

This repository contains the code to generate the dataset for the OpenTofu Registry user interface (backend),
the user interface itself (frontend), as well as the Cloudflare worker powering search.

## Backend

You can run the backend by running `go run ./cmd/generate/main.go` in the [`backend`](backend) directory. This command
has a number of options detailed below.

### General options

| Option                    | Description                                                                                                                                                                                                                                                                                             |
|---------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--skip-update-providers` | Do not update providers, only update modules (if enabled) and regenerate the search index.                                                                                                                                                                                                              |
| `--skip-update-modules`   | Do not update modules, only update providers (if enabled) and regenerate the search index.                                                                                                                                                                                                              |
| `--namespace`             | Limit updates to a namespace.                                                                                                                                                                                                                                                                           |
| `--name`                  | Limit updates to a name. Only works in conjunction with `--namespace`. For providers, this will result in a single provider getting updated. For modules, this will update all target systems under a name.                                                                                             |
| `--target-system`         | Limit updates to a target system for module updates only. Only works in conjunction with `--namespace` and `--name`.                                                                                                                                                                                    |
| `--log-level`             | "Set the log level (trace, debug, info, warn, error).                                                                                                                                                                                                                                                   |
| `--registry-dir`          | Directory to check out the registry in.                                                                                                                                                                                                                                                                 |
| `--vcs-dir`               | Directory to use for checking out providers and modules in.                                                                                                                                                                                                                                             |
| `--commit-parallelism`    | Parallel uploads to use on commit.                                                                                                                                                                                                                                                                      |
| `--tofu-binary-path`      | Temporary: Tofu binary path to use for module schema extraction. This binary must support the `tofu metadata dump` command.                                                                                                                                                                             |
| `--force-regenerate`      | Force regenerating a namespace, name, or target system. This parameter is a comma-separate list consisting of either a namespace, a namespace and a name separated by a `/`, or a namespace, name and target system separated by a `/`. Example: `namespace/name/targetsystem,othernamespace/othername` |

### Storage options

The system supports storing in a local directory or in an S3 bucket. Passing S3 options automatically switches to the
S3 option.

#### Local directory

By default, the storage will write the resulting documents in a local directory. You can change where this directory
is located by passing the `--destination-dir` option.

#### S3 bucket

The S3 storage has the following options:

| Command line option | Environment variable    | Description                                                                                  |
|---------------------|-------------------------|----------------------------------------------------------------------------------------------|
|                     | `AWS_ACCESS_KEY_ID`     | Access key (required).                                                                       |
|                     | `AWS_SECRET_ACCESS_KEY` | Secret key (required).                                                                       |
| `--s3-bucket`       |                         | S3 bucket to use for uploads (required).                                                     |
| `--s3-endpoint`     | `AWS_ENDPOINT_URL_S3`   | S3 endpoint to use for uploads.                                                              |
| `--s3-path-style`   |                         | Use path-style URLs for S3.                                                                  |
| `--s3-ca-cert-file` | `AWS_CA_BUNDLE`         | File containing the CA certificate for the S3 endpoint. Defaults to the system certificates. |
| `--s3-region`       | `AWS_REGION`            | Region to use for S3 uploads. Defaults to auto-detection.                                    |

### Blocklist

This system supports explicitly blocking providers and modules from being indexed by using a blocklist. You can pass a JSON file using the `--blocklist` option of the following format:

```json
{
  "providers": {
    "namespace/name": "Reason why it was blocked (shown to the user).",
    "namespace/": "This entire namespace is blocked."
  },
  "modules": {
    "namespace/name/targetsystem": "Reason why it was blocked (shown to the user).",
    "namespace/name/": "Anything under this name is blocked.",
    "namespace//": "This entire namespace is blocked."
  }
}
```

As the listed providers and modules are in the registry dataset, they will still appear in the UI, but the contents of
the documentation are not shown. This is useful to comply with DMCA take-down requests, erroneous license detection, or
to honor the wishes of the provider/module author.  

### GitHub Actions

This repository also contains a workflow for GitHub Actions. This workflow periodically runs the indexing and requires
the following secrets to be set up:

- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_ENDPOINT_URL_S3`
- `S3_BUCKET`

Additionally, the `Generate (manual)` workflow allows you to manually trigger the generation for a namespace, name, or target system, including an option to force the regeneration from scratch.

## Frontend

In order to run the frontend, enter the [frontend](frontend) directory and run `npm run dev`. You can create a `.env`
file to configure where the generated dataset and search API are:

```env
VITE_DATA_API_URL=http://localhost:8000
```

## Search

The search API is independent of the backend because it requires a PostgreSQL database. It consists of an indexer to
fill the database and a Cloudflare worker to answer search queries.

The indexer reads from the generated search index feed located at the `/search.ndjson` endpoint of the generated data
from the backend and fills the database. You can pass the `--connection-string` option in order to supply it with a 
database connection.

The worker receives all calls to the API from the frontend and processes search requests, passing on any requests it
cannot handle to the R2 (S3-style) bucket containing the dataset. For development, you can set up a `wrangler.toml` in
the following format:

```toml
#:schema node_modules/wrangler/config-schema.json
name = "registry-ui-search"
main = "src/index.ts"
compatibility_date = "2024-08-21"
compatibility_flags = ["nodejs_compat"]
```
