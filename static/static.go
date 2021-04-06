package static

import (
	_ "embed"

	"github.com/evanw/esbuild/pkg/api"
)

//go:generate go run ../generators/style.go

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
	return mainCSS
}

func MakeWASMLoader() []byte {
	result := api.Transform(wasmData, api.TransformOptions{
		MinifyIdentifiers: false,
		MinifySyntax:      true,
		MinifyWhitespace:  true,
	})
	return result.Code
}
