package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:     "mg",
		HelpName: "mg",
		Usage:    "Tools for smugmug",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "smugmug-client-key",
				Required: true,
				Usage:    "smugmug client key",
				EnvVars:  []string{"SMUGMUG_CLIENT_KEY"},
			},
			&cli.StringFlag{
				Name:     "smugmug-client-secret",
				Required: true,
				Usage:    "smugmug client secret",
				EnvVars:  []string{"SMUGMUG_CLIENT_SECRET"},
			},
			&cli.StringFlag{
				Name:     "smugmug-access-token",
				Required: true,
				Usage:    "smugmug access token",
				EnvVars:  []string{"SMUGMUG_ACCESS_TOKEN"},
			},
			&cli.StringFlag{
				Name:     "smugmug-token-secret",
				Required: true,
				Usage:    "smugmug token secret",
				EnvVars:  []string{"SMUGMUG_TOKEN_SECRET"},
			},
			&cli.BoolFlag{
				Name:     "debug",
				Required: false,
				Usage:    "enable debugging",
				Value:    false,
			},
		},
		ExitErrHandler: func(c *cli.Context, err error) {
			if err == nil {
				return
			}
			log.Error().Err(err).Msg(c.App.Name)
		},
		Before: func(c *cli.Context) error {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			zerolog.DurationFieldUnit = time.Millisecond
			zerolog.DurationFieldInteger = false
			log.Logger = log.Output(
				zerolog.ConsoleWriter{
					Out:        c.App.ErrWriter,
					NoColor:    false,
					TimeFormat: time.RFC3339,
				},
			)
			return nil
		},
		Action: func(c *cli.Context) error {
			client, err := smugmug.NewHTTPClient(
				c.String("smugmug-client-key"),
				c.String("smugmug-client-secret"),
				c.String("smugmug-access-token"),
				c.String("smugmug-token-secret"))
			if err != nil {
				return err
			}

			mg, err := smugmug.NewClient(smugmug.WithHTTPClient(client), smugmug.WithHTTPTracing(c.Bool("debug")))
			if err != nil {
				return err
			}

			user, err := mg.User.User(c.Context)
			if err != nil {
				return err
			}

			fmt.Println(user.NickName)

			albums, err := mg.Album.Albums(c.Context, user.NickName)
			if err != nil {
				return err
			}

			for i, album := range albums {
				fmt.Println(" " + album.NiceName)
				if i == 0 {
					images, err := mg.Image.Images(c.Context, album.AlbumKey)
					if err != nil {
						return err
					}
					for _, image := range images {
						fmt.Printf("  %s => '%s'\n", image.FileName, image.Caption)
					}
				}
			}

			return nil
		},
	}
	if err := app.RunContext(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
