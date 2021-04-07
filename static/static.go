package static

import _ "embed"

//go:embed style.min.css
var CSS string

//go:embed owl_wasm.min.js
var WASMLoader []byte

//go:embed owl_wasm_bg.wasm
var WASMBinary []byte

//go:embed fts_exec.min.js
var FTSExecutor string

//go:embed theme_selector.min.js
var ThemeSelector string
