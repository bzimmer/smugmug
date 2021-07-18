package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/smugmug"
)

var mg *smugmug.Client

func main() {
	app := &cli.App{
		Name:     "demo",
		HelpName: "demo",
		Usage:    "Demo for smugmug",
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
				Name:     "verbose",
				Required: false,
				Usage:    "enable verbose logging",
				Value:    false,
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
			level := zerolog.InfoLevel
			if c.Bool("verbose") {
				level = zerolog.DebugLevel
			}
			zerolog.SetGlobalLevel(level)
			zerolog.DurationFieldUnit = time.Millisecond
			zerolog.DurationFieldInteger = false
			log.Logger = log.Output(
				zerolog.ConsoleWriter{
					Out:        c.App.ErrWriter,
					NoColor:    false,
					TimeFormat: time.RFC3339,
				},
			)

			var err error
			client, err := smugmug.NewHTTPClient(
				c.String("smugmug-client-key"),
				c.String("smugmug-client-secret"),
				c.String("smugmug-access-token"),
				c.String("smugmug-token-secret"))
			if err != nil {
				return err
			}

			mg, err = smugmug.NewClient(
				smugmug.WithHTTPClient(client),
				smugmug.WithPretty(c.Bool("debug")),
				smugmug.WithHTTPTracing(c.Bool("debug")))
			if err != nil {
				return err
			}

			return nil
		},
		Commands: []*cli.Command{
			{
				Name: "albums",
				Action: func(c *cli.Context) error {
					user, err := mg.User.User(c.Context)
					if err != nil {
						return err
					}
					i := 0
					page := smugmug.WithPagination(0, 100)
					for {
						albums, pages, err := mg.Album.Albums(c.Context, user.NickName, page)
						if err != nil {
							return err
						}
						for _, album := range albums {
							fmt.Printf("[%04d] [%s] %s\n", i, album.AlbumKey, album.URLName)
							i++
						}

						if pages.NextPage == "" {
							return nil
						}

						fmt.Println("-")
						page = smugmug.WithPagination(pages.Start+pages.Count, 100)
					}
				},
			},
			{
				Name: "nodes",
				Action: func(c *cli.Context) error {
					user, err := mg.User.User(c.Context)
					if err != nil {
						return err
					}
					i := 0
					page := smugmug.WithPagination(0, 100)
					for {
						nodes, pages, err := mg.Node.Search(c.Context, page,
							smugmug.WithSearch(user.URI, c.Args().First()))
						if err != nil {
							return err
						}
						for _, node := range nodes {
							fmt.Printf("[%04d] [%s] %s\n", i, node.NodeID, node.URI)
							i++
						}

						if pages.NextPage == "" {
							return nil
						}

						fmt.Println("-")
						page = smugmug.WithPagination(pages.Start+pages.Count, 100)
					}
				},
			},
			{
				Name: "album",
				Action: func(c *cli.Context) error {
					album, err := mg.Album.Album(c.Context, c.Args().First(), smugmug.WithExpansions("AlbumHighlightImage", "AlbumImages"))
					if err != nil {
						return err
					}
					fmt.Println(album.URLName)
					fmt.Println(" " + album.Expansions.HighlightImage.FileName)
					fmt.Printf(" %03d images\n", len(album.Expansions.Images))
					for _, image := range album.Expansions.Images {
						cover := " "
						if album.Expansions.HighlightImage.FileName == image.FileName {
							cover = "*"
						}
						fmt.Printf("%s  %s | %s |\n", cover, image.FileName, image.Caption)
					}
					return nil
				},
			},
			{
				Name: "search",
				Action: func(c *cli.Context) error {
					user, err := mg.User.User(c.Context)
					if err != nil {
						return err
					}
					log.Info().Str("scope", user.URIs.Node.URI).Msg("search")
					albums, pages, err := mg.Album.Search(c.Context,
						smugmug.WithFilters("Name", "LastUpdated"),
						smugmug.WithSorting(smugmug.DirectionNone, smugmug.MethodLastUpdated),
						smugmug.WithSearch(user.URIs.Node.URI, c.Args().First()),
					)
					if err != nil {
						return err
					}
					fmt.Printf(" albums %d\n", len(albums))
					fmt.Printf(" total %d\n", pages.Total)

					for _, album := range albums {
						fmt.Printf("  %s -- %s\n", album.LastUpdated, album.Name)
					}
					return nil
				},
			},
			{
				Name: "image",
				Action: func(c *cli.Context) error {
					img, err := mg.Image.Image(c.Context, c.Args().First(), smugmug.WithExpansions("ImageSizeDetails"))
					if err != nil {
						return err
					}
					fmt.Printf(" %s %s\n", img.FileName, img.Expansions.ImageSizeDetails.ImageSizeLarge.URL)
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			user, err := mg.User.User(c.Context, smugmug.WithExpansions("UserAlbums", "UserGeoMedia", "CoverImage"))
			if err != nil {
				return err
			}

			fmt.Println(user.NickName)
			fmt.Printf(" %03d albums\n", len(user.Expansions.Albums))

			albums, pages, err := mg.Album.Albums(c.Context, user.NickName)
			if err != nil {
				return err
			}

			fmt.Printf(" pages >> %d/%d\n", pages.Count, pages.Total)
			for i, album := range albums {
				fmt.Printf(" %s | %s |\n", album.URLName, "")
				if i == 0 {
					images, pages, err := mg.Image.Images(c.Context, album.AlbumKey)
					if err != nil {
						return err
					}
					fmt.Printf("  pages >> %d/%d\n", pages.Count, pages.Total)
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
