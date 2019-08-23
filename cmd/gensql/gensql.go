package main

import (
	"database/sql"
	"fmt"
	"os"

	lib "github.com/cncf/apisnoopregexp"
	_ "github.com/lib/pq" // As suggested by lib/pq driver
)

func generateSQL(con *sql.DB) error {
	rows := lib.QuerySQLWithErr(
		con,
		"select distinct operation_id from raw_audit_events where operation_id is not null order by operation_id",
	)
	defer func() { lib.FatalOnError(rows.Close()) }()
	opid := ""
	opids := []string{}
	for rows.Next() {
		lib.FatalOnError(rows.Scan(&opid))
		opids = append(opids, opid)
	}
	lib.FatalOnError(rows.Err())
	for _, opid := range opids {
		rs := lib.QuerySQLWithErr(
			con,
			"select distinct request_uri, verb from raw_audit_events where operation_id = $1",
			opid,
		)
		requesturi := ""
		verb := ""
		sqlRoot := "update audit_events set op_id = '" + opid + "' where ("
		sql := sqlRoot
		args := 0
		for rs.Next() {
			lib.FatalOnError(rs.Scan(&requesturi, &verb))
			sql += "(request_uri = '" + requesturi + "' and verb = '" + verb + "') or "
			args++
			if args == 1000 {
				sql = sql[:len(sql)-4] + ");"
				fmt.Printf("%s\n", sql)
				sql = sqlRoot
				args = 0
			}
		}
		if args > 0 {
			sql = sql[:len(sql)-4] + ");"
			fmt.Printf("%s\n", sql)
		}
		lib.FatalOnError(rs.Err())
		lib.FatalOnError(rs.Close())
	}
	return nil
}

func main() {
	// sudo -u postgres ./gensql
	// psql "host=/var/run/postgresql user=postgres dbname=hh sslmode=disable password=''"
	connectionString := os.Getenv("CONN")
	if connectionString == "" {
		connectionString = lib.ConnStr
	}
	con, err := sql.Open("postgres", connectionString)
	lib.FatalOnError(err)
	lib.FatalOnError(generateSQL(con))
}
