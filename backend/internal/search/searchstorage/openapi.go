package searchstorage

// GetSearchIndex returns a newline-delimited search index suitable for insertion into a database. The records are not
// guaranteed to be in order.
//
// swagger:operation GET /search.ndjson Search GetSearchIndex
// ---
// produces:
// - application/x-ndjson
// responses:
//   '200':
//     description: A newline-delimited search index suitable for insertion into a database. The records are not guaranteed to be in order. Each item is a GeneratedIndexItem.
//     schema:
//       type: array
//       items:
//         $ref: '#/definitions/GeneratedIndexItem'

// SearchResultItem describes a single search result item.
//
// swagger:model SearchResultItem
type SearchResultItem struct {
	// The unique identifier for the result item.
	// required: true
	ID string `json:"id"`
	// The last updated timestamp for the result item.
	// required: true
	LastUpdated string `json:"last_updated"`
	// The type of the result item (e.g., module, provider, datasource etc).
	// required: true
	Type string `json:"type"`
	// The address of the module or provider.
	// required: true
	Addr string `json:"addr"`
	// The version of the module or provider
	// required: true
	Version string `json:"version"`
	// The title of the result item.
	// required: true
	Title string `json:"title"`
	// A brief description of the result item.
	// required: true
	Description string `json:"description"`
	// A map of variables used to generate the link for the result item.
	// required: true
	LinkVariables map[string]string `json:"link_variables"`
	// The number of times the search term matched in this result.
	// required: true
	TermMatchCount string `json:"term_match_count"`
	// The rank of the result in the search results.
	// required: true
	Rank int32 `json:"rank"`
}

// Search returns a list of search results matching the query.
//
// swagger:operation GET /search Search Search
// ---
// produces:
// - application/json
// parameters:
// - description: The search query string. This should be a URL encoded string.
//   in: query
//   name: q
//   required: true
//   type: string
// responses:
//   '200':
//     description: A list of search results matching the query.
//     schema:
//       type: array
//       items:
//         $ref: '#/definitions/SearchResultItem'
//   '400':
//     description: Invalid search query.
// tags:
// - Search
