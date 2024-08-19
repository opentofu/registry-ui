package searchstorage

// GetSearchIndex returns a lunr.js search index.
//
// swagger:operation GET /search.json Search GetSearchIndex
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: A lunr.js search index for modules and providers.
//     schema:
//       type: file
