package main

import (
	"database/sql"
	"fmt"
	"regexp"
	"runtime"
	"sync"

	lib "github.com/cncf/apisnoopregexp"
	_ "github.com/lib/pq" // As suggested by lib/pq driver
)

func rmatchSQL(con *sql.DB) error {
	// Distinct request URIs
	rows := lib.QuerySQLWithErr(
		con,
		fmt.Sprintf(
			"select distinct request_uri from audit_events where op_id is null",
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
	fmt.Printf("URIs: %d\n", len(uris))

	// Distinct regexps
	rows = lib.QuerySQLWithErr(
		con,
		fmt.Sprintf(
			"select distinct regexp from api_operations",
		),
	)
	defer func() { lib.FatalOnError(rows.Close()) }()
	re := ""
	res := []*regexp.Regexp{}
	rmap := make(map[*regexp.Regexp]string)
	for rows.Next() {
		lib.FatalOnError(rows.Scan(&re))
		rex := regexp.MustCompile(re)
		res = append(res, rex)
		rmap[rex] = re
	}
	lib.FatalOnError(rows.Err())
	fmt.Printf("RegExps: %d\n", len(res))

	// Now find all matches
	thrN := runtime.NumCPU()
	runtime.GOMAXPROCS(thrN)
	ch := make(chan struct{})
	matches := make(map[string][]*regexp.Regexp)
	mut := &sync.Mutex{}
	nThreads := 0
	for _, u := range uris {
		go func(c chan struct{}, uri string) {
			ms := []*regexp.Regexp{}
			for _, re := range res {
				m := re.MatchString(uri)
				if m {
					ms = append(ms, re)
				}
			}
			mut.Lock()
			matches[uri] = ms
			mut.Unlock()
			c <- struct{}{}
		}(ch, u)
		nThreads++
		if nThreads == thrN {
			<-ch
			nThreads--
		}
	}
	for nThreads > 0 {
		<-ch
		nThreads--
	}

	// Matching analysis
	hist := make(map[int]int)
	for _, m := range matches {
		l := len(m)
		v, ok := hist[l]
		if !ok {
			hist[l] = 0
		} else {
			hist[l] = v + 1
		}
	}
	fmt.Printf("Matches data: %+v\n", hist)

	// Now update uris
	nThreads = 0
	ch = make(chan struct{})
	for u, m := range matches {
		go func(c chan struct{}, uri string, ms []*regexp.Regexp) {
			rs := lib.QuerySQLWithErr(
				con,
				fmt.Sprintf(
					"select distinct verb from audit_events where op_id is null and request_uri = $1",
				),
				uri,
			)
			verb := ""
			verbs := []string{}
			for rs.Next() {
				lib.FatalOnError(rs.Scan(&verb))
				verbs = append(verbs, verb)
			}
			lib.FatalOnError(rs.Err())
			lib.FatalOnError(rs.Close())
			c <- struct{}{}
		}(ch, u, m)
		nThreads++
		if nThreads == thrN {
			<-ch
			nThreads--
		}
	}
	for nThreads > 0 {
		<-ch
		nThreads--
	}
	fmt.Printf("Done\n")
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
