package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/opentofu/registry-ui/internal/indexstorage/filesystemstorage"
	"github.com/opentofu/registry-ui/internal/license"
	"github.com/opentofu/registry-ui/internal/providerindex"
	"github.com/opentofu/registry-ui/internal/providerindex/providerindexstorage"
	"github.com/opentofu/registry-ui/internal/registry/github"
	"github.com/opentofu/registry-ui/internal/registry/provider"

	_ "github.com/mattn/go-sqlite3"
)

func mockDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./fake.db")
	if err != nil {
		panic(err)
	}

	schema, err := os.ReadFile("../search/pg-indexer/schema_sqlite.sql")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		panic(err)
	}

	return db
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Starting version bump process for modules and providers")

	licensesFile := ""

	//moduleDataDir := flag.String("module-data", "../modules", "Directory containing the module data")
	//moduleNamespace := flag.String("module-namespace", "", "Which module namespace to limit the command to")
	providerDataDir := flag.String("provider-data", "../providers", "Directory containing the provider data")
	providerNamespace := flag.String("provider-namespace", "", "Which provider namespace to limit the command to")
	flag.StringVar(&licensesFile, "licenses-file", licensesFile, "JSON file containing a list of approved licenses to include when indexing. (required)")

	flag.Parse()

	ctx := context.Background()

	// HACK IN DB FOR NOW
	db := mockDB()
	defer db.Close()

	token, err := github.EnvAuthToken()
	if err != nil {
		logger.Error("Initialization Error", slog.Any("err", err))
		os.Exit(1)
	}
	ghClient := github.NewClient(ctx, logger, token)

	approvedLicenses, err := readLicensesFile(ctx, licensesFile)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
	licenseDetector := license.NewDetector(approvedLicenses)

	providers, err := provider.ListProviders(*providerDataDir, *providerNamespace, logger, ghClient)
	if err != nil {
		logger.Error("Failed to list providers", slog.Any("err", err))
		os.Exit(1)
	}
	//log *slog.Logger, ctx context.Context, providers provider.List, licenseDetector license.Detector, destination providerindexstorage.API, searchAPI search.API, workDir string
	fss, err := filesystemstorage.New("./generated")
	if err != nil {
		logger.Error("Failed to cam72cam", slog.Any("err", err))
		os.Exit(1)
	}
	destination, _ := providerindexstorage.New(fss)
	err = providerindex.GenerateDocumentation(logger, ctx, providers, licenseDetector, destination, db, "./work")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("Completed version bump process for modules and providers")
}

func readLicensesFile(_ context.Context, licensesFile string) ([]string, error) {
	if licensesFile == "" {
		return nil, fmt.Errorf("the --licenses-file parameter is required")
	}

	licensesFileContent, err := os.ReadFile(licensesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read licenses file: %w", err)
	}

	re := regexp.MustCompile(`//.*?\n`)
	licensesFileContent = re.ReplaceAll(licensesFileContent, nil)

	var approvedLicenses []string
	if err := json.Unmarshal(licensesFileContent, &approvedLicenses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal licenses file: %w", err)
	}
	if len(approvedLicenses) == 0 {
		return nil, fmt.Errorf("the licenses file contains no licenses")
	}
	return approvedLicenses, nil
}
