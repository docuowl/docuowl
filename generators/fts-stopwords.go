//+build ignore

package main

import (
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	files, err := ioutil.ReadDir("lang")
	if err != nil {
		panic(err)
	}

	stopWords := map[string][]string{}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		path := "./lang/" + f.Name()
		contents, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		rawContents := strings.Split(string(contents), "\n")
		newContents := make([]string, 0, len(contents))
		for _, l := range rawContents {
			if strings.HasPrefix(l, "#") {
				continue
			}
			if strings.TrimSpace(l) == "" {
				continue
			}
			newContents = append(newContents, "\""+l+"\"")
		}
		fileName := filepath.Base(path)
		lang := strings.TrimSuffix(strings.TrimPrefix(fileName, "stopwords-"), ".txt")

		stopWords[lang] = newContents
	}

	newFile := []string{
		"package lang",
		"",
		"// Code generated by go:generate. DO NOT EDIT.",
		"",
		"var StopWords = map[string][]string {",
	}

	for lang, strs := range stopWords {
		newFile = append(newFile, fmt.Sprintf(`"%s": []string{%s},`, lang, strings.Join(strs, ", ")))
	}

	newFile = append(newFile, "}", "")
	source := []byte(strings.Join(newFile, "\n"))
	source, err = format.Source(source)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("lang/stopwords.go", source, os.ModePerm)
	if err != nil {
		panic(err)
	}
}
