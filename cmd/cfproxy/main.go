package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "cfproxy",
		Usage: "A 3rd-party implementation of cloudflared bastion mode",
		Commands: cli.Commands{
			{
				Name:    "connect",
				Aliases: []string{"c"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "target",
						Aliases:  []string{"T"},
						Usage:    "Target address to connect to",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "local",
						Aliases:  []string{"L"},
						Usage:    "Local address to bind to",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "client-id",
						Aliases: []string{"id"},
						Usage:   "Client ID for Cloudflare Access",
					},
					&cli.StringFlag{
						Name:    "client-secret",
						Aliases: []string{"secret"},
						Usage:   "Client Secret for Cloudflare Access",
					},
				},
				Action: func(context *cli.Context) error {
					logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
					connect(&connectOptions{
						target:       context.String("target"),
						local:        context.String("local"),
						logger:       &logger,
						clientId:     context.String("client-id"),
						clientSecret: context.String("client-secret"),
					})
					return nil
				},
			},
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Action: func(context *cli.Context) error {
					fmt.Println("SERVE")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
