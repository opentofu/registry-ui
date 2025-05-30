# generate-registry runs `go generate` to build the registry files and saves them to /tmp/registry
# since it's a high number of files, we are limiting to opentofu namespace.
# If more providers are needed, tweak the arguments on this command
generate-registry:
	@cd "$(CURDIR)/backend" && go generate ./... && go run ./cmd/generate/ --licenses-file ../licenses.json --destination-dir /tmp/registry --namespace opentofu --name ad

# remove-provider removes a provider from the registry
# Example: make remove-provider PROVIDER=hashicorp/aws
# Example with flags: make remove-provider PROVIDER=hashicorp/aws REMOVE_FLAGS="--dry-run --version v1.0.0"
remove-provider:
	@cd "$(CURDIR)/backend" && go run ./cmd/remove/ $(REMOVE_FLAGS) provider $(PROVIDER)

# load-registry feed the data from /tmp/registry into the local R2 bucket (search/worker/.wrangler/state/r2) folder
load-registry:
	@cd "$(CURDIR)/search/worker" && npm run feed-data

# index-search downloads search data from api.opentofu.org and feeds that data into the postgres database used for searching
index-search:
	@cd "$(CURDIR)/search/pg-indexer" && PG_CONNECTION_STRING=postgres://postgres:secret@localhost:5432/postgres?sslmode=disable go run . 

# after docker-compose us running, run this command to feed data into the application
feed-data:
	make generate-registry
	make load-registry
	make index-search
