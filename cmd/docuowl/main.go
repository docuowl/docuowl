package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/heyvito/docuowl/fs"
	"github.com/heyvito/docuowl/fts"
	"github.com/heyvito/docuowl/parts"
	"github.com/heyvito/docuowl/static"
	"github.com/heyvito/docuowl/watch"
)

const version = "0.2.3"

func main() {

	app := cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "watch",
				Usage: "Serves OUTPUT in PORT, and automatically reloads the page when changes are detected.",
				Value: false,
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "When --watch is used, defines in which port to serve output",
				Value: 8000,
			},
			&cli.StringFlag{
				Name:  "lang",
				Usage: "Language in which your documentation is written",
				Value: "en",
			},
			&cli.StringFlag{
				Name:     "input",
				Usage:    "Where to look for input files",
				Value:    "",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "output",
				Usage:    "Where to output compiled documentation",
				Value:    "",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "no-fts",
				Usage: "Disables FTS facilities",
				Value: false,
			},
		},
		Action: func(c *cli.Context) error {
			var (
				doWatch = c.Bool("watch")
				noFTS   = c.Bool("no-fts")
				port    = c.Int("port")
				lang    = c.String("lang")
				input   = c.String("input")
				output  = c.String("output")
			)
			inputPath, err := filepath.Abs(input)
			if err != nil {
				log.Fatalf("Error processing input path %s: %s", input, err)
			}

			outputPath, err := filepath.Abs(output)
			if err != nil {
				log.Fatalf("Error processing output path %s: %s", input, err)
			}

			if doWatch {
				if err = render(lang, inputPath, outputPath, noFTS); err != nil {
					log.Printf("Error executing initial render: %s", err)
					log.Println("Will retry on next update")
				}
				w := watch.New(inputPath, outputPath, port, func() error {
					return render(lang, inputPath, outputPath, noFTS)
				})
				log.Printf("Listening on 127.0.0.1:%d", port)
				if err = w.Run(); err != nil {
					log.Fatalf("%s", err)
				}
			} else {
				log.Println("Rendering contents...")
				err := render(lang, inputPath, outputPath, noFTS)
				if err != nil {
					log.Fatalf("%s", err)
				}
			}
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func render(lang, input, output string, noFTS bool) error {
	tree, err := fs.Walk(input)
	if err != nil {
		return fmt.Errorf("error scanning %s: %w", input, err)
	}

	ftsInst := fts.New(lang)

	sidebar := parts.MakeSidebar(tree, version, noFTS)
	rendered := parts.RenderItems(tree, ftsInst)
	index, err := ftsInst.Serialize()
	if err != nil {
		return fmt.Errorf("error serialising FTS index: %w", err)
	}
	head := parts.MakeHead(index, noFTS)
	if err = os.MkdirAll(output, os.ModePerm); err != nil {
		return fmt.Errorf("error creating directories for %s: %w", output, err)
	}
	if err = os.WriteFile(output+"/index.html", []byte(head+sidebar+rendered), os.ModePerm); err != nil {
		return fmt.Errorf("error writing %s: %w", output+"/index.html", err)
	}

	if noFTS {
		return nil
	}
	if err = os.WriteFile(output+"/owl_wasm.js", static.WASMLoader, os.ModePerm); err != nil {
		return fmt.Errorf("error writing %s: %s", output+"/owl_wasm.js", err)
	}

	if err = os.WriteFile(output+"/owl_wasm_bg.wasm", static.WASMBinary, os.ModePerm); err != nil {
		return fmt.Errorf("error writing %s: %s", output+"/owl_wasm_bg.wasm", err)
	}
	return nil
}
