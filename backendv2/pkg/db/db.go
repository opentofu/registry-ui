// Package db handles db connections and migrations
// For construction of a database client, please use the [config] package
package db

import (
	_ "github.com/jackc/pgx/v5/stdlib"
)
