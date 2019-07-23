package main

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"time"

	_ "github.com/lib/pq" // As suggested by lib/pq driver
)

func fatalOnError(err error) {
	if err != nil {
		tm := time.Now()
		fmt.Printf("Error(time=%+v):\nError: '%s'\nStacktrace:\n%s\n", tm, err.Error(), string(debug.Stack()))
		fmt.Fprintf(os.Stderr, "Error(time=%+v):\nError: '%s'\nStacktrace:\n", tm, err.Error())
		panic("stacktrace")
	}
}

func fatalf(f string, a ...interface{}) {
	fatalOnError(fmt.Errorf(f, a...))
}

func queryOut(query string, args ...interface{}) {
	fmt.Printf("%s\n", query)
	if len(args) > 0 {
		s := ""
		for vi, vv := range args {
			switch v := vv.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128, string, bool, time.Time:
				s += fmt.Sprintf("%d:%+v ", vi+1, v)
			default:
				s += fmt.Sprintf("%d:%+v ", vi+1, reflect.ValueOf(vv).Elem())
			}
		}
		fmt.Printf("[%s]\n", s)
	}
}

func querySQL(con *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	return con.Query(query, args...)
}

func querySQLWithErr(con *sql.DB, query string, args ...interface{}) *sql.Rows {
	var (
		res *sql.Rows
		err error
	)
	res, err = querySQL(con, query, args...)
	if err != nil {
		queryOut(query, args...)
	}
	fatalOnError(err)
	return res
}

func generateSQL(con *sql.DB) error {
	rows := querySQLWithErr(
		con,
		fmt.Sprintf(
			"select distinct op_id from audit_events where op_id is not null order by op_id",
		),
	)
	defer func() { fatalOnError(rows.Close()) }()
	opid := ""
	opids := []string{}
	for rows.Next() {
		fatalOnError(rows.Scan(&opid))
		opids = append(opids, opid)
	}
	fatalOnError(rows.Err())
	for _, opid := range opids {
		rs := querySQLWithErr(
			con,
			fmt.Sprintf(
				"select distinct request_uri, verb from audit_events where op_id = $1",
			),
			opid,
		)
		requesturi := ""
		verb := ""
		sqlRoot := "update audit_events set op_id = '" + opid + "' where ("
		sql := sqlRoot
		args := 0
		for rs.Next() {
			fatalOnError(rs.Scan(&requesturi, &verb))
			sql += "(request_uri = '" + requesturi + "' and verb = '" + verb + "') or "
			args++
			if args == 500 {
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
		fatalOnError(rs.Err())
		fatalOnError(rs.Close())
	}
	return nil
}

func main() {
	// sudo -u postgres ./gensql
	// psql "host=/var/run/postgresql user=postgres dbname=hh sslmode=disable password=''"
	connectionString := "client_encoding=UTF8 sslmode=disable host=/var/run/postgresql port=5432 dbname=hh user=postgres password=''"
	con, err := sql.Open("postgres", connectionString)
	fatalOnError(err)
	fatalOnError(generateSQL(con))
}
