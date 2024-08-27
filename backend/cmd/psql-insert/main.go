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
	// TODO: Do config properly using a lib, not just flags
	// TODO: Break out into smaller files/packages
	var connString string
	flag.StringVar(&connString, "connection-string", "", "Postgres connection string")

	var batchSize int
	flag.IntVar(&batchSize, "batch-size", 1000, "Batch size for inserts")
	flag.Parse()

	if connString == "" {
		log.Fatal("connection string is required")
	}

	if batchSize <= 0 {
		log.Fatal("batch size must be greater than 0")
	}

	err := downloadSearchMetaIndex()
	if err != nil {
		log.Fatal(err)
	}

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
	if mostRecentJob.InProgress() {
		log.Println("Import already in progress, exiting")
		os.Exit(0)
	}

	file, err := os.Open("./search.ndjson")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	header, err := readSearchHeader(scanner.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	headerTime, err := time.Parse(time.RFC3339Nano, header.Header.LastUpdated)
	if err != nil {
		log.Fatal(err)
	}

	shouldRun := false
	if mostRecentJob == nil {
		shouldRun = true
	} else if mostRecentJob.CompletedAt.Valid && mostRecentJob.CompletedAt.Time.Before(headerTime) {
		shouldRun = true
	}

	if !shouldRun {
		log.Println("No new data to import, exiting")
		log.Printf("\tLast import completed at %s\n", mostRecentJob.CompletedAt.Time.Format(time.RFC3339Nano))
		log.Printf("\tLast data update was at %s\n", headerTime.Format(time.RFC3339Nano))
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
		tx.Rollback()
		log.Fatal(err)
	}

	for len(batchItems) > 0 {
		toInsert := make([]SearchIndexItem, 0, batchSize)
		toDelete := make([]SearchIndexItem, 0, batchSize)

		for _, item := range batchItems {
			if item.Type == "add" {
				toInsert = append(toInsert, item)
				handled++
			} else if item.Type == "deletion" {
				toDelete = append(toDelete, item)
				handled++
			}
		}

		if err := insertItems(tx, toInsert); err != nil {
			tx.Rollback()
			log.Fatal(err)
		}

		if err := deleteItems(tx, toDelete); err != nil {
			tx.Rollback()
			log.Fatal(err)
		}

		batchItems, err = readItems(scanner, batchSize)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}

		log.Printf("Imported %d items\n", handled)
	}

	tx.Commit()
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

	if header.Header.LastUpdated == "" {
		return nil, fmt.Errorf("last_updated field is required")
	}

	return &header, nil
}

func (i *ImportJob) InProgress() bool {
	if i == nil {
		return false
	}
	return !i.CompletedAt.Valid
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
	// Start a new job
	id := 0
	err := db.QueryRow("INSERT INTO import_jobs (created_at) VALUES (NOW()) RETURNING id").Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func completeJob(db *sql.DB, id int) error {
	// set completed_at and successful to true for the job with id
	_, err := db.Exec("UPDATE import_jobs SET completed_at = NOW(), successful = true WHERE id = $1", id)
	return err
}

func downloadSearchMetaIndex() error {
	log.Println("Downloading search.ndjson")
	out, err := os.Create("./search.ndjson")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get("https://api.opentofu.org/search.ndjson")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	// stream the data to disk
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Downloaded search.ndjson")

	return err
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
	args := make([]interface{}, 0, len(items)*7) // 7 fields + 1 for timestamp

	for i, item := range items {
		linkVarsJSON, err := json.Marshal(item.Addition.Link)
		if err != nil {
			return fmt.Errorf("failed to marshal link variables: %w", err)
		}

		// Build the placeholder for each row
		placeholderIndex := i * 7 // 7 fields per row
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, NOW())",
			placeholderIndex+1, placeholderIndex+2, placeholderIndex+3,
			placeholderIndex+4, placeholderIndex+5, placeholderIndex+6, placeholderIndex+7))

		args = append(args,
			item.Addition.ID,
			item.Addition.Type,
			item.Addition.Addr,
			item.Addition.Version,
			item.Addition.Title,
			item.Addition.Description,
			linkVarsJSON,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO entities (id, type, addr, version, title, description, link_variables, last_updated)
		VALUES %s
		ON CONFLICT (id) DO UPDATE 
		SET type = EXCLUDED.type,
		    addr = EXCLUDED.addr,
		    version = EXCLUDED.version,
		    title = EXCLUDED.title,
		    description = EXCLUDED.description,
		    link_variables = EXCLUDED.link_variables,
		    last_updated = NOW()
	`, strings.Join(values, ","))

	// Execute the query with all the arguments
	_, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	return nil
}
