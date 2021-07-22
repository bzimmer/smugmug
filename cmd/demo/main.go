package main

import (
	"context"
	"encoding/json"
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
					user, err := mg.User.AuthUser(c.Context)
					if err != nil {
						return err
					}
					i := 0
					return mg.Album.AlbumsIter(c.Context, user.NickName, func(album *smugmug.Album) error {
						fmt.Printf("[%04d] [%s] %s\n", i, album.AlbumKey, album.URLName)
						i++
						return nil
					})
				},
			},
			{
				Name: "children",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "images",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					img := c.Bool("images")
					enc := json.NewEncoder(c.App.Writer)
					f := func(node *smugmug.Node) error {
						switch node.Type {
						case "Album":
							log.Info().
								Str("nodeID", node.NodeID).
								Str("albumKey", node.Album.AlbumKey).
								Str("type", node.Type).
								Str("name", node.Name).
								Int("imageCount", node.Album.ImageCount).
								Msg("images")
							if img {
								return mg.Image.ImagesIter(c.Context, node.Album.AlbumKey, func(image *smugmug.Image) error {
									if image.Caption == "" {
										return nil
									}
									return enc.Encode(map[string]interface{}{
										"filename":  image.FileName,
										"caption":   image.Caption,
										"latitude":  image.Latitude,
										"longitude": image.Longitude,
									})
								})
							}
						case "Folder":
							fallthrough
						default:
							log.Info().Str("nodeID", node.NodeID).Str("name", node.Name).Str("type", node.Type).Msg("children")
						}
						return nil
					}
					nodeID := c.Args().First()
					return mg.Node.ChildrenIter(c.Context, nodeID, f, smugmug.WithExpansions("Album", "FolderByID", "HighlightImage", "User"))
				},
			},
			{
				Name: "nodes",
				Action: func(c *cli.Context) error {
					user, err := mg.User.AuthUser(c.Context)
					if err != nil {
						return err
					}
					i := 0
					return mg.Node.SearchIter(
						c.Context,
						func(node *smugmug.Node) error {
							fmt.Printf("[%04d] [%s] %s %s\n", i, node.NodeID, node.URI, node.Name)
							i++
							return nil
						},
						smugmug.WithExpansions("ParentNode"),
						smugmug.WithSearch(user.URI, c.Args().First()))
				},
			},
			{
				Name: "album",
				Action: func(c *cli.Context) error {
					album, err := mg.Album.Album(c.Context, c.Args().First(), smugmug.WithExpansions("AlbumHighlightImage"))
					if err != nil {
						return err
					}
					fmt.Println(album.URLName)
					fmt.Println(" " + album.HighlightImage.FileName)
					fmt.Printf(" %03d images\n", album.ImageCount)

					f := func(image *smugmug.Image) error {
						cover := " "
						if album.HighlightImage.FileName == image.FileName {
							cover = "*"
						}
						fmt.Printf("%s  %s | %s %s |\n", cover, image.FileName, image.ImageKey, image.Caption)
						return nil
					}
					return mg.Image.ImagesIter(c.Context, album.AlbumKey, f)
				},
			},
			{
				Name: "search",
				Action: func(c *cli.Context) error {
					user, err := mg.User.AuthUser(c.Context)
					if err != nil {
						return err
					}
					log.Info().Str("scope", user.URIs.Node.URI).Msg("search")
					albums, pages, err := mg.Album.Search(c.Context,
						smugmug.WithFilters("Name", "LastUpdated"),
						smugmug.WithSorting("", "Last Updated"),
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
					fmt.Printf(" %s %s\n", img.FileName, img.ImageSizeDetails.ImageSizeLarge.URL)
					return nil
				},
			},
			{
				Name: "walk",
				Action: func(c *cli.Context) error {
					return mg.Node.Walk(c.Context, c.Args().First(), func(node *smugmug.Node) error {
						switch node.Type {
						case "Album":
							log.Info().
								Str("nodeID", node.NodeID).
								Str("albumKey", node.Album.AlbumKey).
								Str("type", node.Type).
								Str("name", node.Name).
								Int("imageCount", node.Album.ImageCount).
								Msg("images")
						case "Folder":
							fallthrough
						case "Node":
							log.Info().
								Str("nodeID", node.NodeID).
								Str("name", node.Name).
								Str("type", node.Type).
								Msg("children")
						}
						return nil
					}, smugmug.WithExpansions("Album"))
				},
			},
		},
		Action: func(c *cli.Context) error {
			user, err := mg.User.AuthUser(c.Context)
			if err != nil {
				return err
			}

			fmt.Println(user.NickName)

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
