package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "cfproxy",
		Usage: "The 3rd-party implementation of cloudflared bastion mode",
		Commands: cli.Commands{
			{
				Name:    "connect",
				Aliases: []string{"c"},
				Action: func(context *cli.Context) error {
					fmt.Println("CONNECT")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
