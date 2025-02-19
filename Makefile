# generate-registry runs `go generate` to build the registry files and saves them to /tmp/registry
generate-registry:
	@cd "$(CURDIR)/backend" && go run ./cmd/generate/ --licenses-file ../licenses.json --destination-dir /tmp/registry

# load-registry feed the data from /tmp/registry into the local R2 bucket (search/worker/.wrangler/state/r2) folder
load-registry:
	@cd "$(CURDIR)/search/worker" && npm run feed-data

# index-search downloads search data from api.opentofu.org and feeds that data into the postgres database used for searching
index-search:
	@cd "$(CURDIR)/search/pg-indexer" && PG_CONNECTION_STRING=postgres://postgres:secret@localhost:5432/postgres?sslmode=disable go run . 
