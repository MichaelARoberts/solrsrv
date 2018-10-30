package main

import (
	"fmt"
	"github.com/vanng822/go-solr/solr"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/urfave/cli.v1/altsrc"
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

	app.Before = altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewYamlSourceFromFlagFunc("config"))

	app.Action = func(c *cli.Context) error {
		si, _ := solr.NewSolrInterface(fmt.Sprintf(
			"http://%s:%d", c.String("solr.host"),
			c.Int("solr.port")),
			c.String("solr.collection"))
		fmt.Printf("Solr Client Instance: %v", si)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}

	fmt.Println("solrsrv will now exit.")
}
