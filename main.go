package main

import (
	// "errors"
	"fmt"
	// "github.com/vanng822/go-solr/solr"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/urfave/cli.v1/altsrc"
	"net/http"
	"os"
)

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
			Name:  "solr.collection",
			Value: "default",
			Usage: "set the default collection to use",
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
		// si, _ := solr.NewSolrInterface(fmt.Sprintf(
		// 	"http://%s:%d", c.String("solr.host"),
		// 	c.Int("solr.port")),
		// 	c.String("solr.collection"))
		// fmt.Printf("Solr Client Instance: %v", si)

		http.HandleFunc("/complete", func(w http.ResponseWriter, req *http.Request) {
			switch req.Method {
			case http.MethodGet:
				if err := req.ParseForm(); err != nil {
					panic(err)
				}
				q := req.Form["q"]
				query := ""
				if len(q) > 0 {
					query = q[0]
				}
				// w.Write([]byte(fmt.Sprintf("req: %v\n", req)))
				w.Write([]byte(fmt.Sprintf("Query: %s\n", query)))
			default:
				msg := fmt.Sprintf("Invalid Method: Expecting GET, received %v\n", req.Method)
				code := http.StatusMethodNotAllowed
				w.WriteHeader(code)
				if len, err := w.Write([]byte(msg)); err != nil {
					fmt.Printf("resp len: %v; err: %v\n", len, err)
				}
			}
		})

		if err := http.ListenAndServe(":8080", nil); err != nil {
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
