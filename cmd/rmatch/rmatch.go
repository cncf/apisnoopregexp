package main

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sync"

	lib "github.com/cncf/apisnoopregexp"
	_ "github.com/lib/pq" // As suggested by lib/pq driver
)

func rmatchSQL(con *sql.DB) error {
	// Set to true to have verbose output
	dbg := os.Getenv("DBG") != ""
	// Distinct request URIs
	rows := lib.QuerySQLWithErr(
		con,
		"select distinct request_uri from raw_audit_event where operation_id is null",
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
		"select distinct regex from api_operation order by regex asc",
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
	fmt.Printf("Regexps: %d\n", len(res))

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
			// Bottleneck, but still goes very fast
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
	if os.Getenv("ANALYSIS") != "" {
		hist := make(map[int]int)
		for _, m := range matches {
			l := len(m)
			v, ok := hist[l]
			if !ok {
				hist[l] = 1
			} else {
				hist[l] = v + 1
			}
		}
		fmt.Printf("Matches data (ideally there should be no misses (0:0) and no multiple matches (>=2:0), only (1:N):\n%+v\n", hist)
	}

	// Now update op_id
	nThreads = 0
	ch = make(chan struct{})
	updated := int64(0)
	for u, m := range matches {
		go func(c chan struct{}, uri string, ms []*regexp.Regexp) {
			lms := len(ms)
			if lms < 1 {
				if dbg {
					fmt.Printf("INFO: No matches: uri:%s\n", uri)
				}
				c <- struct{}{}
				return
			}
			rs := lib.QuerySQLWithErr(
				con,
				"select distinct event_verb from raw_audit_event where operation_id is null and request_uri = $1",
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
			if len(verbs) < 1 {
				fmt.Printf("ERROR: No verbs: uri:%s matches:%d\n", uri, lms)
			}
			for _, verb := range verbs {
				method := ""
				if verb == lib.Get || verb == lib.List || verb == lib.Proxy {
					method = lib.Get
				} else if verb == lib.Patch {
					method = lib.Patch
				} else if verb == lib.Update {
					method = lib.Put
				} else if verb == lib.Create {
					method = lib.Post
				} else if verb == lib.Delete || verb == lib.Deletecollection {
					method = lib.Delete
				} else if verb == lib.Watch || verb == lib.Watchlist {
					method = lib.Watch
				} else {
					fmt.Printf("WARNING: unknown verb:%s for uri:%s\n", verb, uri)
					continue
				}
				//fmt.Printf("uri:%s: verb:%s -> method:%s\n", uri, verb, method)
				aids := []string{}
				aid := ""
				pre := ""
				lre := 0
				for _, ma := range ms {
					// This key must exist, algorithm ensures this, otherwise it would panic here
					sma := rmap[ma]
					lsma := len(sma)
					// There should be at most one
					rs2 := lib.QuerySQLWithErr(
						con,
						"select operation_id from api_operation_material where http_method = $1 and regex = $2",
						method,
						sma,
					)
					id := ""
					ids := []string{}
					for rs2.Next() {
						lib.FatalOnError(rs2.Scan(&id))
						ids = append(ids, id)
						aids = append(aids, id)
						if lsma > lre {
							if dbg && lre > 0 {
								fmt.Printf("DEBUG: picking longer(%d/%s):%s instead of shorter(%d/%s):%s\n", lsma, id, sma, lre, aid, pre)
							}
							aid = id
							lre = lsma
							pre = sma
						}
					}
					lib.FatalOnError(rs2.Err())
					lib.FatalOnError(rs2.Close())
					//fmt.Printf("verb:%s method:%s regexp:%s -> ids:%+v\n", verb, method, sma, ids)
					if len(ids) > 1 {
						fmt.Printf("ERROR: Multiple method/regexp matches: uri:%s: verb:%s method:%s regexp:%s -> ids:%+v\n", uri, verb, method, sma, ids)
					}
				}
				//fmt.Printf("verb:%s method:%s -> ids:%+v\n", verb, method, aids)
				la := len(aids)
				if la < 1 {
					if dbg {
						fmt.Printf("INFO: No IDs found: uri:%s verb:%s method:%s regexps:%+v\n", uri, verb, method, ms)
					}
				} else {
					if la > 1 {
						fmt.Printf("WARNING: Multiple IDs found: uri:%s verb:%s method:%s regexps:%+v -> ids:%+v, picking longest regexp: (%d/%s):%s\n", uri, verb, method, ms, aids, lre, aid, pre)
					}
					// which ms is longer?
					// what is it's index
					// use aids[longer_ms_index]
					rt := lib.ExecSQLWithErr(
						con,
						"update raw_audit_event set operation_id = $1 where request_uri = $2 and event_verb = $3",
						aids[0],
						uri,
						verb,
					)
					cnt, e := rt.RowsAffected()
					lib.FatalOnError(e)
					if cnt > 0 {
						mut.Lock()
						updated += cnt
						mut.Unlock()
					}
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
	fmt.Printf("Done, updated %d rows\n", updated)
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
	lib.FatalOnError(rmatchSQL(con))
}
