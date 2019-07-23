package apisnoopregexp

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"time"

	_ "github.com/lib/pq" // As suggested by lib/pq driver
)

// ConnStr - postgres connection string
const ConnStr string = "client_encoding=UTF8 sslmode=disable host=/var/run/postgresql port=5432 dbname=hh user=postgres password=''"

// FatalOnError - fail on error
func FatalOnError(err error) {
	if err != nil {
		tm := time.Now()
		fmt.Printf("Error(time=%+v):\nError: '%s'\nStacktrace:\n%s\n", tm, err.Error(), string(debug.Stack()))
		fmt.Fprintf(os.Stderr, "Error(time=%+v):\nError: '%s'\nStacktrace:\n", tm, err.Error())
		panic("stacktrace")
	}
}

// Fatalf - fail on a given string
func Fatalf(f string, a ...interface{}) {
	FatalOnError(fmt.Errorf(f, a...))
}

// QueryOut - output query and its params (using reflection if needed)
func QueryOut(query string, args ...interface{}) {
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

// QuerySQL - execute SQl query
func QuerySQL(con *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	return con.Query(query, args...)
}

// QuerySQLWithErr - execute SQL query and eventually fail
func QuerySQLWithErr(con *sql.DB, query string, args ...interface{}) *sql.Rows {
	var (
		res *sql.Rows
		err error
	)
	res, err = QuerySQL(con, query, args...)
	if err != nil {
		QueryOut(query, args...)
	}
	FatalOnError(err)
	return res
}
