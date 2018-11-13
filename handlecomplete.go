package main

import (
	"encoding/json"
	"fmt"
	"github.com/vanng822/go-solr/solr"
	"gopkg.in/urfave/cli.v1"
	"net/http"
)

func handleComplete(c *cli.Context, si *solr.SolrInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			handleCompleteGet(w, req, c, si)
		default:
			msg := fmt.Sprintf("Invalid Method: Expecting GET, received %v\n", req.Method)
			code := http.StatusMethodNotAllowed
			w.WriteHeader(code)
			if len, err := w.Write([]byte(msg)); err != nil {
				fmt.Printf("resp len: %v; err: %v\n", len, err)
			}
		}
	}
}

func handleCompleteGet(w http.ResponseWriter, req *http.Request,
	c *cli.Context, si *solr.SolrInterface,
) {
	if err := req.ParseForm(); err != nil {
		panic(err)
	}
	q := req.Form["q"]
	if c.Bool("dryrun") {
		query := ""
		if len(q) > 0 {
			query = q[0]
		}
		// w.Write([]byte(fmt.Sprintf("req: %v\n", req)))
		w.Write([]byte(fmt.Sprintf("Query: %s\n", query)))
		return
	}
	query := "*"
	if len(q) > 0 {
		query = q[0]
	}
	q := solr.NewQuery()
	q.AddParam("q", fmt.Sprintf("manu_autocomplete:%s", query))
	q.AddParam("wt", "json")
	q.AddParam("fl", "manu")
	dbg := &DebugParser{}
	res, err := si.Search(q).Result(dbg)
	if err != nil {
		panic(err)
	}
	docs := res.Results.Docs
	// resultsMap := map[string]bool{}
	results := []string{}
	last := ""
	for _, doc := range docs {
		res := doc.Get("manu").(string)
		if res != last {
			results = append(results, doc.Get("manu").(string))
		}
		// resultsMap[doc.Get("manu").(string)] = true
	}
	// results := []string{}
	// for res := range resultsMap {
	// 	results = append(results, res)
	// }
	resJson, _ := json.Marshal(results)
	if len, err := w.Write(resJson); err != nil {
		fmt.Printf("resp len: %v; err: %v\n", len, err)
	}
}
