package frontmatter

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/heyvito/docuowl/slug"
)

type Meta struct {
	Title string `yaml:"Title"`
	ID    string `yaml:"ID"`
}

var ErrUnexpectedEOF = fmt.Errorf("unexpected eof")

func ExtractFromLines(lines []string) (*Meta, []string, error) {
	if len(lines) < 3 {
		return nil, lines, nil
	}

	if lines[0] != "---" {
		return nil, lines, nil
	}

	var frontmatterLines []string
	var file []string
	var ok bool

	for i, l := range lines {
		if i == 0 {
			continue
		}
		if l == "---" {
			file = lines[i+1:]
			ok = true
			break
		}
		frontmatterLines = append(frontmatterLines, l)
	}

	if !ok {
		return nil, nil, ErrUnexpectedEOF
	}

	var meta Meta
	err := yaml.Unmarshal([]byte(strings.Join(frontmatterLines, "\n")), &meta)
	if err != nil {
		return nil, nil, err
	}

	if meta.ID == "" {
		meta.ID = meta.Title
	}
	meta.ID = slug.Slugify(meta.ID)

	allBlank := true
	for _, l := range file {
		if strings.TrimSpace(l) != "" {
			allBlank = false
			break
		}
	}
	if allBlank {
		file = []string{}
	}
	return &meta, file, nil
}

func ExtractFromFile(path string) (*Meta, []string, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	return ExtractFromLines(strings.Split(string(fileBytes), "\n"))
}
