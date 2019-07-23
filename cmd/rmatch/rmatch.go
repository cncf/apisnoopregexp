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
	// Set to true to have verbose output
	dbg := false
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
			if dbg && len(ms) < 1 {
				fmt.Printf("No matches: uri:%s -> verbs:%+v, matches:%d\n", uri, verbs, len(ms))
			}
			if len(verbs) < 1 {
				fmt.Printf("No verbs: uri:%s -> verbs:%+v, matches:%d\n", uri, verbs, len(ms))
			}
			for _, verb := range verbs {
				method := ""
				if verb == lib.Get || verb == "list" || verb == "proxy" {
					method = lib.Get
				} else if verb == lib.Patch {
					method = lib.Patch
				} else if verb == "update" {
					method = "put"
				} else if verb == "create" {
					method = "post"
				} else if verb == lib.Delete || verb == "deletecollection" {
					method = lib.Delete
				} else if verb == lib.Watch || verb == "watchlist" {
					method = lib.Watch
				} else {
					fmt.Printf("WARNING: unknown verb:%s for uri:%s\n", verb, uri)
					continue
				}
				//fmt.Printf("uri:%s: verb:%s -> method:%s\n", uri, verb, method)
				aids := []string{}
				for _, ma := range ms {
					rs2 := lib.QuerySQLWithErr(
						con,
						fmt.Sprintf(
							"select id from api_operations where method = $1 and regexp = $2",
						),
						method,
						rmap[ma],
					)
					id := ""
					ids := []string{}
					for rs2.Next() {
						lib.FatalOnError(rs2.Scan(&id))
						ids = append(ids, id)
						aids = append(aids, id)
					}
					lib.FatalOnError(rs2.Err())
					lib.FatalOnError(rs2.Close())
					//fmt.Printf("uri:%s: verb:%s method:%s regexp:%s -> ids:%+v\n", uri, verb, method, rmap[ma], ids)
					//fmt.Printf("verb:%s method:%s regexp:%s -> ids:%+v\n", verb, method, rmap[ma], ids)
				}
				//fmt.Printf("verb:%s method:%s -> ids:%+v\n", verb, method, aids)
				if dbg && len(aids) < 1 {
					fmt.Printf("uri:%s verb:%s method:%s regexps:%+v -> ids:%+v\n", uri, verb, method, ms, aids)
				}
			}
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
