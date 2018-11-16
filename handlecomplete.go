package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vanng822/go-solr/solr"
	"gopkg.in/urfave/cli.v1"
	"log"
	"net/http"
)

func handleComplete(c *cli.Context, si *solr.SolrInterface, stderr *log.Logger,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			if err := handleCompleteGet(w, req, c, si); err != nil {
				stderr.Println(err)
			}
		default:
			msg := fmt.Sprintf("Invalid Method: Expecting GET, received %v\n", req.Method)
			code := http.StatusMethodNotAllowed
			w.WriteHeader(code)
			if len, err := w.Write([]byte(msg)); err != nil {
				stderr.Printf("resp len: %v; err: %v\n",
					len, err)
			}
		}
	}
}

func handleCompleteGet(w http.ResponseWriter, req *http.Request,
	c *cli.Context, si *solr.SolrInterface,
) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	q := req.Form["q"]
	if c.Bool("dryrun") {
		query := ""
		if len(q) > 0 {
			query = q[0]
		}
		// w.Write([]byte(fmt.Sprintf("req: %v\n", req)))
		w.Write([]byte(fmt.Sprintf("Query: %s\n", query)))
		return nil
	}
	query := "*"
	if len(q) > 0 {
		query = q[0]
	}
	sq := solr.NewQuery()
	sq.AddParam("q", fmt.Sprintf("name_ac:%s", query))
	sq.AddParam("wt", "json")
	sq.AddParam("fl", "name")
	dbg := &DebugParser{}
	res, err := si.Search(sq).Result(dbg)
	if err != nil {
		return err
	}
	docs := res.Results.Docs
	// resultsMap := map[string]bool{}
	results := []string{}
	last := ""
	for _, doc := range docs {
		res := doc.Get("name").(string)
		if res != last {
			results = append(results, doc.Get("name").(string))
		}
		// resultsMap[doc.Get("name").(string)] = true
	}
	// results := []string{}
	// for res := range resultsMap {
	// 	results = append(results, res)
	// }
	resJson, _ := json.Marshal(results)
	if len, err := w.Write(resJson); err != nil {
		er := errors.New(fmt.Sprintf("resp len: %v; err: %v\n",
			len, err))
		return er
	}
	return nil
}
