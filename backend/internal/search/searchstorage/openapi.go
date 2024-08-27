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
