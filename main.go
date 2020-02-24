package main

import (
	"fmt"
	"github.com/rs/cors"
	"github.com/vanng822/go-solr/solr"
	"gopkg.in/urfave/cli.v2"
	"gopkg.in/urfave/cli.v2/altsrc"
	"log"
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
		altsrc.NewBoolFlag(cli.BoolFlag{
			Name:  "solr.tls",
			Usage: "use https to connect to solr",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "addr",
			Value: "",
			Usage: "determine where solrsrv should listen (see http.Server.Addr for syntax)",
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
	stderr := log.New(os.Stderr, "", 0)

	app.Action = func(c *cli.Context) error {
		var si *solr.SolrInterface
		if c.Bool("dryrun") {
			fmt.Println("Dry run...")
		} else {
			solrFormat := "http://%s:%d/solr"
			if c.Bool("solr.tls") {
				solrFormat = "https://%s:%d/solr"
			}
			si, _ = solr.NewSolrInterface(fmt.Sprintf(
				solrFormat, c.String("solr.host"),
				c.Int("solr.port")),
				c.String("solr.core"))
			fmt.Printf("Solr Client Instance: %v", si)
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/complete", handleComplete(c, si, stderr))
		handler := cors.AllowAll().Handler(mux)

		if err := http.ListenAndServe(c.String("addr"), handler); err != nil {
			stderr.Println(err)
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		stderr.Println(err)
	}

	fmt.Println("solrsrv will now exit.")
}
