//+build ignore

package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/wellington/go-libsass"
)

func main() {
	input, err := os.Open("../static/style.scss")
	if err != nil {
		panic(err)
	}
	var buffer bytes.Buffer

	currentDir, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	comp, err := libsass.New(&buffer, input,
		libsass.OutputStyle(libsass.COMPRESSED_STYLE),
		libsass.IncludePaths([]string{currentDir}))

	if err != nil {
		panic(err)
	}

	if err := comp.Run(); err != nil {
		panic(err)
	}

	if err = os.WriteFile("../static/style.css", buffer.Bytes(), os.ModePerm); err != nil {
		panic(err)
	}
}
