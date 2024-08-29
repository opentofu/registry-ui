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

/*
   /search:
       get:
           operationId: Search
           parameters:
               - description: The search query string. This should be a URL encoded string.
                 in: query
                 name: q
                 required: true
                 type: string
           produces:
               - application/json
           responses:
               "200":
                   description: A list of search results matching the query.
                   schema:
                       type: array
                       items:
                           $ref: '#/definitions/SearchResultItem'
               "400":
                   description: Invalid search query.
           tags:
               - Search
*/

// SearchResultItem describes a single search result item.
//
// swagger:model SearchResultItem
type SearchResultItem struct {
	// The unique identifier for the result item.
	ID string `json:"id"`
	// The last updated timestamp for the result item.
	LastUpdated string `json:"last_updated"`
	// The type of the result item (e.g., module, provider, datasource etc).
	Type string `json:"type"`
	// The address of the module or provider.
	Addr string `json:"addr"`
	// The version of the module or provider
	Version string `json:"version"`
	// The title of the result item.
	Title string `json:"title"`
	// A brief description of the result item.
	Description string `json:"description"`
	// A map of variables used to generate the link for the result item.
	LinkVariables map[string]string `json:"link"`
	// The number of times the search term matched in this result.
	TermMatchCount string `json:"term_match_count"`
	// The rank of the result in the search results.
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
