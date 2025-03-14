package providerindex

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
	"github.com/opentofu/registry-ui/internal/search/searchtypes"
)

/*func providerVersions(db *sql.DB, providerDetails *providertypes.Provider) ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT version FROM entities WHERE addr = ? AND id LIKE 'providers/%'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return versions, nil
}*/

/*
	func deleteItems(tx *sql.Tx, providerDetails *providertypes.Provider, providerVersionDetails providertypes.ProviderVersion) error {
		// Create a query with the necessary number of placeholders
		query := "DELETE FROM entities WHERE type in ('resource', 'datasource', 'function', 'provider') AND addr = ? AND version = ?"
		_, err := tx.Exec(query, providerDetails.Addr.String(), providerVersionDetails.ID)
		if err != nil {
			return fmt.Errorf("failed to delete items: %w", err)
		}

		return nil
	}
*/
func deleteItems(tx *sql.Tx, items []searchtypes.IndexItem) error {
	if len(items) == 0 {
		return nil
	}

	ids := make([]interface{}, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}

	// Create a query with the necessary number of placeholders
	query := "DELETE FROM entities WHERE id = $1"
	for _, id := range ids {
		_, err := tx.Exec(query, id)
		if err != nil {
			return fmt.Errorf("failed to delete items: %w", err)
		}
	}

	return nil
}

func insertItems(tx *sql.Tx, providerDetails *providertypes.Provider, providerVersionDetails providertypes.ProviderVersion) error {
	items := indexProviderVersion(providerDetails, providerVersionDetails)

	if len(items) == 0 {
		return nil
	}

	if err := deleteItems(tx, items); err != nil {
		return err
	}

	values := make([]string, 0, len(items))
	args := make([]interface{}, 0, len(items)*8)

	for i, item := range items {
		linkVarsJSON, err := json.Marshal(item.LinkVariables)
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

		args = append(args, item.ID,
			item.Type,
			item.Addr,
			item.Version,
			item.Title,
			item.Description,
			linkVarsJSON,
			time.Now(), //item.LastUpdated,
			item.Popularity,
			item.Warnings,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO entities (id, type, addr, version, title, description, link_variables, last_updated, popularity, warnings)
		VALUES %s
	`, strings.Join(values, ","))

	// Execute the query with all the arguments
	_, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	return nil
}

func indexProviderVersion(providerDetails *providertypes.Provider, providerVersionDetails providertypes.ProviderVersion) (results []searchtypes.IndexItem) {
	providerAddr := providerDetails.Addr
	version := providerVersionDetails.ProviderVersionDescriptor.ID
	popularity := providerDetails.Popularity
	/* TODO cam72cam if providerAddr.ToRepositoryAddr() == providerDetails.ForkOf.ToRepositoryAddr() {
		// If the non-canonical repo address matches where we forked from, we take the popularity of the upstream.
		popularity = providerDetails.UpstreamPopularity
	}*/
	providerItem := searchtypes.IndexItem{
		ID:          searchtypes.IndexID(indexPrefix + "/" + providerAddr.String()),
		Type:        searchtypes.IndexTypeProvider,
		Addr:        providerAddr.String(),
		Version:     string(version),
		Title:       providerAddr.Name,
		Description: "",
		LinkVariables: map[string]string{
			"namespace": providerAddr.Namespace,
			"name":      providerAddr.Name,
			"version":   string(version),
		},
		ParentID:   "",
		Popularity: popularity,
		Warnings:   len(providerDetails.Warnings),
	}

	results = append(results, providerItem)

	for _, item := range []struct {
		typeName  string
		indexType searchtypes.IndexType
		items     []providertypes.ProviderDocItem
	}{
		{
			"resource",
			searchtypes.IndexTypeProviderResource,
			providerVersionDetails.Docs.Resources,
		},
		{
			"datasource",
			searchtypes.IndexTypeProviderDatasource,
			providerVersionDetails.Docs.DataSources,
		},
		{
			"function",
			searchtypes.IndexTypeProviderFunction,
			providerVersionDetails.Docs.Functions,
		},
	} {
		for _, docItem := range item.items {
			title := docItem.Title
			results = append(results, searchtypes.IndexItem{
				ID:          searchtypes.IndexID(indexPrefix + "/" + providerAddr.String() + "/" + item.typeName + "s/" + string(docItem.Name)),
				Type:        item.indexType,
				Addr:        providerAddr.String(),
				Version:     string(version),
				Title:       title,
				Description: docItem.Description,
				LinkVariables: map[string]string{
					"namespace": providerAddr.Namespace,
					"name":      providerAddr.Name,
					"version":   string(version),
					"id":        string(docItem.Name),
				},
				ParentID:   providerItem.ID,
				Popularity: popularity,
				Warnings:   len(providerDetails.Warnings),
			})
		}
	}
	return results
}
