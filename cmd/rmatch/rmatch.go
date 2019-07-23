package main

import (
	"database/sql"
	"fmt"

	lib "github.com/cncf/apisnoopregexp"
	_ "github.com/lib/pq" // As suggested by lib/pq driver
)

func rmatchSQL(con *sql.DB) error {
	rows := lib.QuerySQLWithErr(
		con,
		fmt.Sprintf(
			"select distinct request_uri from audit_events",
		),
	)
	defer func() { lib.FatalOnError(rows.Close()) }()
	uri := ""
	uris := []string{}
	for rows.Next() {
		lib.FatalOnError(rows.Scan(&uri))
		uris = append(uris, uri)
	}
	lib.FatalOnError(rows.Err())
	fmt.Printf("URIs(%d): %+v\n", len(uris), uris)
	return nil
}

func main() {
	// sudo -u postgres ./gensql
	// psql "host=/var/run/postgresql user=postgres dbname=hh sslmode=disable password=''"
	connectionString := lib.ConnStr
	con, err := sql.Open("postgres", connectionString)
	lib.FatalOnError(err)
	lib.FatalOnError(rmatchSQL(con))
}
