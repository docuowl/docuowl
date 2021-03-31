package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/heyvito/docuowl/frontmatter"
)

type EntityKind int

const (
	EntityKindSection EntityKind = iota + 1
	EntityKindGroup
)

type Entity interface {
	Kind() EntityKind
	Section() *Section
	Group() *Group
	Meta() *frontmatter.Meta
	CompoundID() string
	ParentEntity() Entity
}

func normalizeCompoundID(ids []string) string {
	for i, j := 0, len(ids)-1; i < j; i, j = i+1, j-1 {
		ids[i], ids[j] = ids[j], ids[i]
	}
	return strings.Join(ids, "-")
}

type Section struct {
	Metadata           *frontmatter.Meta
	Content, SideNotes []string
	HasSideNotes       bool
	Parent             *Group
}

func (s *Section) Kind() EntityKind {
	return EntityKindSection
}

func (s *Section) Section() *Section {
	return s
}

func (s *Section) Group() *Group {
	return nil
}

func (s *Section) Meta() *frontmatter.Meta {
	return s.Metadata
}

func (s *Section) ParentEntity() Entity {
	return s.Parent
}

func (s *Section) CompoundID() string {
	var items []string
	var obj Entity = s
	for {
		if isEntityNil(obj) {
			break
		}
		items = append(items, obj.Meta().ID)
		obj = obj.ParentEntity()
	}
	return normalizeCompoundID(items)
}

type Group struct {
	Metadata *frontmatter.Meta
	Children []Entity
	Parent   *Group
	Content  *Section
}

func (g *Group) Kind() EntityKind {
	return EntityKindGroup
}

func (g *Group) Section() *Section {
	return nil
}

func (g *Group) Group() *Group {
	return g
}

func (g *Group) Meta() *frontmatter.Meta {
	return g.Metadata
}

func (g *Group) ParentEntity() Entity {
	return g.Parent
}

func (g *Group) CompoundID() string {
	var items []string
	var obj Entity = g
	for {
		if isEntityNil(obj) {
			break
		}
		items = append(items, obj.Meta().ID)
		obj = obj.ParentEntity()
	}
	return normalizeCompoundID(items)
}
func isEntityNil(ent Entity) bool {
	return ent == nil || reflect.ValueOf(ent).Kind() == reflect.Ptr && reflect.ValueOf(ent).IsNil()
}

func isGroup(matches []string) bool {
	for _, n := range matches {
		if filepath.Base(n) == "meta.md" {
			return true
		}
	}
	return false
}

func isContent(matches []string) bool {
	for _, n := range matches {
		if filepath.Base(n) == "content.md" {
			return true
		}
	}
	return false
}

func isSideNote(matches []string) bool {
	for _, n := range matches {
		if filepath.Base(n) == "sidenotes.md" {
			return true
		}
	}
	return false
}

func walk(root string, parent *Group) ([]Entity, error) {
	var result []Entity
	matches, err := filepath.Glob(filepath.Join(root, "*"))
	if err != nil {
		return nil, err
	}

	if isGroup(matches) {
		meta, content, err := frontmatter.ExtractFromFile(filepath.Join(root, "meta.md"))
		if err != nil {
			return nil, err
		}

		g := &Group{
			Metadata: meta,
			Parent:   parent,
		}
		var partial []Entity
		for _, m := range matches {
			if strings.HasSuffix(m, ".md") {
				continue
			}
			p, err := walk(m, g)
			if err != nil {
				return nil, err
			}
			partial = append(partial, p...)
		}
		g.Children = partial

		if len(content) > 0 {
			g.Content = &Section{Content: content}
		}

		return append(result, g), nil
	}

	if isContent(matches) {
		contentPath := filepath.Join(root, "content.md")
		contentBytes, err := os.ReadFile(contentPath)
		if err != nil {
			return nil, err
		}

		meta, content, err := frontmatter.ExtractFromLines(strings.Split(string(contentBytes), "\n"))
		if err == frontmatter.ErrUnexpectedEOF {
			return nil, fmt.Errorf("%s: Could not find frontmatter terminator", contentPath)
		} else if err != nil {
			return nil, err
		}

		if meta == nil {
			return nil, fmt.Errorf("%s: Content files must have frontmatter", contentPath)
		}

		section := &Section{
			Metadata: meta,
			Content:  content,
			Parent:   parent,
		}

		if isSideNote(matches) {
			rawSides, err := os.ReadFile(filepath.Join(root, "sidenotes.md"))
			if err != nil {
				return nil, err
			}

			section.HasSideNotes = true
			section.SideNotes = strings.Split(string(rawSides), "\n")
		}

		result = append(result, section)
	} else {
		var subdirs []string
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path == root {
				return nil
			}
			if info.IsDir() {
				if strings.HasPrefix(filepath.Base(path), ".") {
					return nil
				}
				subdirs = append(subdirs, path)
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

		for _, s := range subdirs {
			entries, err := walk(s, parent)
			if err != nil {
				return nil, err
			}
			result = append(result, entries...)
		}
	}

	return result, nil
}

func Walk(root string) ([]Entity, error) {
	return walk(root, nil)
}
