package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/wellington/go-libsass"
)

func generateCSS() error {
	input, err := os.Open("./static/scss/style.scss")
	if err != nil {
		return err
	}
	var buffer bytes.Buffer

	currentDir, err := filepath.Abs("./static/scss")
	if err != nil {
		return err
	}
	comp, err := libsass.New(&buffer, input,
		libsass.OutputStyle(libsass.COMPRESSED_STYLE),
		libsass.IncludePaths([]string{currentDir}))

	if err != nil {
		return err
	}

	if err := comp.Run(); err != nil {
		return err
	}

	if err = os.WriteFile("./static/style.min.css", buffer.Bytes(), os.ModePerm); err != nil {
		return err
	}

	return nil
}

func compileJS(from, to string, mangle bool) error {
	input, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	result := api.Transform(string(input), api.TransformOptions{
		MinifyIdentifiers: mangle,
		MinifySyntax:      true,
		MinifyWhitespace:  true,
	})

	return os.WriteFile(to, result.Code, os.ModePerm)
}

func generateJS() error {
	funcs := []struct {
		from   string
		mangle bool
	}{
		{from: "fts_exec.js", mangle: false},
		{from: "theme_selector.js", mangle: true},
		{from: "toggle_menu.js", mangle: true},
		{from: "owl_wasm.js", mangle: false},
	}
	for _, meta := range funcs {
		minJS := strings.TrimSuffix(meta.from, ".js") + ".min.js"

		if err := compileJS("./static/js/"+meta.from, "./static/"+minJS, meta.mangle); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var err error
	if err = generateCSS(); err != nil {
		panic(err)
	}
	if err = generateJS(); err != nil {
		panic(err)
	}
}
