package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func main() {
	connString := os.Getenv("PG_CONNECTION_STRING")
	var batchSize int
	flag.IntVar(&batchSize, "batch-size", 1000, "Batch size for inserts")
	flag.Parse()

	if connString == "" {
		log.Fatal("Please set the PG_CONNECTION_STRING environment variable.")
	}

	if batchSize <= 0 {
		log.Fatal("The batch size must be greater than 0.")
	}

	body, err := downloadSearchMetaIndex()
	if err != nil {
		log.Fatal(err)
	}
	defer body.Close()

	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mostRecentJob, err := getMostRecentJob(db)
	if err != nil {
		log.Fatal(err)
	}

	// If the most recent job is in progress, exit
	// TODO: Handle this in a better way
	if mostRecentJob.InProgress() {
		log.Println("Import already in progress, exiting")
		os.Exit(0)
	}

	scanner := bufio.NewScanner(body)
	if !scanner.Scan() {
		log.Fatal("no data in search index")
	}

	header, err := readSearchHeader(scanner.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	shouldRun := false
	if mostRecentJob == nil {
		shouldRun = true
	} else if mostRecentJob.CreatedAt.Time.Before(header.Header.LastUpdated) {
		shouldRun = true
	}

	if !shouldRun {
		log.Println("No new data to import, exiting")
		log.Printf("\tLast import completed at %s\n", mostRecentJob.CreatedAt.Time.Format(time.RFC3339Nano))
		log.Printf("\tLast data update was at %s\n", header.Header.LastUpdated)
		os.Exit(0)
	}

	fmt.Println("Importing data")
	jobID, err := startJob(db)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	handled := 0
	batchItems, err := readItems(scanner, batchSize)
	if err != nil {
		_ = tx.Rollback()
		log.Fatal(err)
	}

	for len(batchItems) > 0 {
		toInsert := make([]SearchIndexItem, 0, batchSize)
		toDelete := make([]SearchIndexItem, 0, batchSize)

		for _, item := range batchItems {
			if item.Type == "add" {
				toInsert = append(toInsert, item)
				handled++
			} else if item.Type == "delete" {
				toDelete = append(toDelete, item)
				handled++
			} else {
				log.Printf("Skipping unknown item type: %s\n", item.Type)
			}
		}

		if err := insertItems(tx, toInsert); err != nil {
			_ = tx.Rollback()
			log.Fatal(err)
		}

		if err := deleteItems(tx, toDelete); err != nil {
			_ = tx.Rollback()
			log.Fatal(err)
		}

		batchItems, err = readItems(scanner, batchSize)
		if err != nil {
			_ = tx.Rollback()
			log.Fatal(err)
		}

		log.Printf("Imported %d items\n", handled)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}
	log.Printf("Complete: Handled %d items\n", handled)

	err = completeJob(db, jobID)
	if err != nil {
		log.Fatal(err)
	}
}

func readItems(scanner *bufio.Scanner, batchSize int) ([]SearchIndexItem, error) {
	batchItems := make([]SearchIndexItem, 0, batchSize)

	for i := 0; i < batchSize; i++ {
		if !scanner.Scan() {
			break
		}

		var item SearchIndexItem
		err := json.Unmarshal(scanner.Bytes(), &item)
		if err != nil {
			return nil, err
		}
		batchItems = append(batchItems, item)
	}

	return batchItems, nil
}

func readSearchHeader(data []byte) (*SearchHeader, error) {
	var header SearchHeader
	err := json.Unmarshal(data, &header)
	if err != nil {
		return nil, err
	}

	if header.ItemType != "header" {
		return nil, fmt.Errorf("expected header type, got %s", header.ItemType)
	}

	return &header, nil
}

func (i *ImportJob) InProgress() bool {
	return i != nil && !i.CompletedAt.Valid
}

func getMostRecentJob(db *sql.DB) (*ImportJob, error) {
	// Check if we need to do a new import by getting the latest import_job item from the db
	rows, err := db.Query("SELECT * FROM import_jobs WHERE successful=true ORDER BY id DESC LIMIT 1 ")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	var id int
	var createdAt sql.NullTime
	var completedAt sql.NullTime
	var successful bool

	for rows.Next() {
		if err := rows.Scan(&id, &createdAt, &completedAt, &successful); err != nil {
			return nil, err
		}
	}

	// if there are no rows, return a new ImportJob with no completedAt
	if id == 0 {
		return nil, nil
	}

	return &ImportJob{
		ID:          id,
		CreatedAt:   createdAt,
		CompletedAt: completedAt,
		Successful:  successful,
	}, nil
}

func startJob(db *sql.DB) (int, error) {
	now := time.Now()

	id := 0
	err := db.QueryRow("INSERT INTO import_jobs (created_at) VALUES ($1) RETURNING id", now).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func completeJob(db *sql.DB, id int) error {
	now := time.Now()
	_, err := db.Exec("UPDATE import_jobs SET completed_at = $1, successful = true WHERE id = $2", now, id)
	return err
}

func downloadSearchMetaIndex() (io.ReadCloser, error) {
	// Get the data
	resp, err := http.Get("https://api.opentofu.org/registry/docs/search.ndjson")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to download search index: %s", resp.Status)
	}

	return resp.Body, nil
}

func deleteItems(tx *sql.Tx, items []SearchIndexItem) error {
	if len(items) == 0 {
		return nil
	}

	ids := make([]interface{}, len(items))
	for i, item := range items {
		ids[i] = item.Deletion.ID
	}

	// Create a query with the necessary number of placeholders
	query := "DELETE FROM entities WHERE id = ANY($1)"
	_, err := tx.Exec(query, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("failed to delete items: %w", err)
	}

	return nil
}

func insertItems(tx *sql.Tx, items []SearchIndexItem) error {
	if len(items) == 0 {
		return nil
	}

	values := make([]string, 0, len(items))
	args := make([]interface{}, 0, len(items)*8)

	for i, item := range items {
		linkVarsJSON, err := json.Marshal(item.Addition.Link)
		if err != nil {
			return fmt.Errorf("failed to marshal link variables: %w", err)
		}

		// Build the placeholder for each row, this is to be used to construct the sql query and not for the actual values
		// it's okay to use sprintf here as no values are actually being injected, we're just building the query
		placeholderIndex := i * 10 // 10 fields per row
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			placeholderIndex+1, placeholderIndex+2, placeholderIndex+3,
			placeholderIndex+4, placeholderIndex+5, placeholderIndex+6,
			placeholderIndex+7, placeholderIndex+8, placeholderIndex+9,
			placeholderIndex+10))

		args = append(args,
			item.Addition.ID,
			item.Addition.Type,
			item.Addition.Addr,
			item.Addition.Version,
			item.Addition.Title,
			item.Addition.Description,
			linkVarsJSON,
			item.Addition.LastUpdated,
			item.Addition.Popularity,
			item.Addition.Warnings,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO entities (id, type, addr, version, title, description, link_variables, last_updated, popularity, warnings)
		VALUES %s
		ON CONFLICT (id) DO UPDATE
		SET type = EXCLUDED.type,
				addr = EXCLUDED.addr,
				version = EXCLUDED.version,
				title = EXCLUDED.title,
				description = EXCLUDED.description,
				link_variables = EXCLUDED.link_variables,
				last_updated = EXCLUDED.last_updated,
				popularity = EXCLUDED.popularity,
				warnings = EXCLUDED.warnings
	`, strings.Join(values, ","))

	// Execute the query with all the arguments
	_, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	return nil
}
