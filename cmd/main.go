package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/heyvito/docuowl/fs"
	fts2 "github.com/heyvito/docuowl/fts"
	"github.com/heyvito/docuowl/parts"
	"github.com/heyvito/docuowl/static"
	watch2 "github.com/heyvito/docuowl/watch"
)

const version = "0.2"

func main() {
	var (
		watch  bool
		noFTS  bool
		port   int
		lang   string
		input  string
		output string
	)
	flag.BoolVar(&watch, "watch", false, "Serves OUTPUT in PORT, and automatically reloads the page when changes are detected.")
	flag.IntVar(&port, "port", 8000, "When --watch is used, defines in which port to serve output")
	flag.StringVar(&lang, "lang", "en", "Language in which your documentation is written")
	flag.StringVar(&input, "input", "", "Where to look for input files")
	flag.StringVar(&output, "output", "", "Where to output compiled documentation")
	flag.BoolVar(&noFTS, "no-fts", false, "Disables FTS facilities")
	flag.Parse()

	log.Printf("Docuowl v%s", version)

	if input == "" {
		_, _ = fmt.Fprintln(os.Stderr, "ERROR: -input is required.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if output == "" {
		_, _ = fmt.Fprintln(os.Stderr, "ERROR: -output is required.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	inputPath, err := filepath.Abs(input)
	if err != nil {
		log.Fatalf("Error processing input path %s: %s", input, err)
	}

	outputPath, err := filepath.Abs(output)
	if err != nil {
		log.Fatalf("Error processing output path %s: %s", input, err)
	}

	if watch {
		if err = render(lang, inputPath, outputPath, noFTS); err != nil {
			log.Printf("Error executing initial render: %s", err)
			log.Println("Will retry on next update")
		}
		w := watch2.New(inputPath, outputPath, port, func() error {
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
}

func render(lang, input, output string, noFTS bool) error {
	tree, err := fs.Walk(input)
	if err != nil {
		return fmt.Errorf("error scanning %s: %w", input, err)
	}

	fts := fts2.New(lang)

	sidebar := parts.MakeSidebar(tree, version, noFTS)
	rendered := parts.RenderItems(tree, fts)
	index, err := fts.Serialize()
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
	if err = os.WriteFile(output+"/owl_wasm.js", static.MakeWASMLoader(), os.ModePerm); err != nil {
		return fmt.Errorf("error writing %s: %s", output+"/owl_wasm.js", err)
	}

	if err = os.WriteFile(output+"/owl_wasm_bg.wasm", static.WASMBinary, os.ModePerm); err != nil {
		return fmt.Errorf("error writing %s: %s", output+"/owl_wasm_bg.wasm", err)
	}
	return nil
}
