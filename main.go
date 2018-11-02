package main

import (
	"encoding/json"
	// "errors"
	"fmt"
	"github.com/rs/cors"
	"github.com/vanng822/go-solr/solr"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/urfave/cli.v1/altsrc"
	"net/http"
	"os"
)

type DebugParser struct{}

func (parser *DebugParser) Parse(resp *[]byte) (*solr.SolrResult, error) {
	fmt.Printf("%v", resp)
	srp := &solr.StandardResultParser{}
	return srp.Parse(resp)
}

func main() {
	app := cli.NewApp()
	app.Name = "solrsrv"
	app.Usage = "Expose a simple REST api for web clients to interact " +
		"with a Solr instance."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "solrsrv.yaml",
			Usage: "Load configuration from YAML `FILE`",
		},
		cli.BoolFlag{
			Name:  "dryrun, d",
			Usage: "Don't actually connect/search Solr",
		},
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "solr.host",
			Value: "127.0.0.1",
			Usage: "set host of solr instance backing solrsrv",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "solr.port",
			Value: 8983,
			Usage: "set port of solr instance backing solrsrv",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "solr.core",
			Value: "default",
			Usage: "set the core to use",
		}),
	}

	// app.Before = altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewYamlSourceFromFlagFunc("config"))
	app.Before = func(c *cli.Context) error {
		if _, err := os.Stat(c.String("config")); os.IsNotExist(err) {
			// return errors.New("foobar")
			return nil
		}
		_, err := altsrc.NewYamlSourceFromFlagFunc("config")(c)
		return err
	}

	app.Action = func(c *cli.Context) error {
		var si *solr.SolrInterface
		if c.Bool("dryrun") {
			fmt.Println("Dry run...")
		} else {
			si, _ = solr.NewSolrInterface(fmt.Sprintf(
				"http://%s:%d/solr", c.String("solr.host"),
				c.Int("solr.port")),
				c.String("solr.core"))
			fmt.Printf("Solr Client Instance: %v", si)
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/complete", func(w http.ResponseWriter, req *http.Request) {
			switch req.Method {
			case http.MethodGet:
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
				} else {
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
					resultsMap := map[string]bool{}
					for _, doc := range docs {
						// results = append(results, doc.Get("manu").(string))
						resultsMap[doc.Get("manu").(string)] = true
					}
					results := []string{}
					for res := range resultsMap {
						results = append(results, res)
					}
					resJson, _ := json.Marshal(results)
					if len, err := w.Write(resJson); err != nil {
						fmt.Printf("resp len: %v; err: %v\n", len, err)
					}

				}
			default:
				msg := fmt.Sprintf("Invalid Method: Expecting GET, received %v\n", req.Method)
				code := http.StatusMethodNotAllowed
				w.WriteHeader(code)
				if len, err := w.Write([]byte(msg)); err != nil {
					fmt.Printf("resp len: %v; err: %v\n", len, err)
				}
			}
		})
		handler := cors.Default().Handler(mux)

		if err := http.ListenAndServe(":80", handler); err != nil {
			panic(err)
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}

	fmt.Println("solrsrv will now exit.")
}
