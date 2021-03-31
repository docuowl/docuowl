package static

import (
	_ "embed"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
)

//go:embed reset.css
var resetCSS string

//go:embed style.css
var mainCSS string

//go:embed owl_wasm.js
var wasmData string

//go:embed owl_wasm_bg.wasm
var WASMBinary []byte

//go:embed fts_exec.js
var FTSExecutor string

func MakeExecutor() string {
	result := api.Transform(FTSExecutor, api.TransformOptions{
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		MinifyWhitespace:  true,
	})
	return string(result.Code)
}

func MakeStyles() string {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	b, err := m.Bytes("text/css", []byte(resetCSS+mainCSS))
	if err != nil {
		// Should not happen.
		panic(err)
	}
	return string(b)
}

func MakeWASMLoader() []byte {
	result := api.Transform(wasmData, api.TransformOptions{
		MinifyIdentifiers: false,
		MinifySyntax:      true,
		MinifyWhitespace:  true,
	})
	return result.Code
}
