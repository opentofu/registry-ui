# Search worker

To setup your local environment for this project:

1. Copy the `.example.dev.vars` to a new `.dev.vars` file in the same folder.
2. Postgres Search data: 

- Go to `../pg-indexer` folder and run `PG_CONNECTION_STRING=postgres://postgres:secret@localhost:5432/postgres?sslmode=disable go run . ` to populate the postgres search.

3. R2 data from providers:

Go to `../../backend` folder and run:

```
go run ./cmd/generate/ --licenses-file ../licenses.json --tofu-binary-path=$(which tofu) --destination-dir /tmp/registry --namespace <namespace>
```

to generate for a namespace and for a full load use:

```
go run ./cmd/generate/ --licenses-file ../licenses.json --tofu-binary-path=$(which tofu)$ --destination-dir /tmp/registry
```

This will create a folder in /tmp/registry with all the registry data.

To load these files into our R2 local bucket, use the script:

```
npm run feed-data
```
