package parts

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/heyvito/docuowl/fs"
	"github.com/heyvito/docuowl/fts"
	"github.com/heyvito/docuowl/markdown"
	"github.com/heyvito/docuowl/static"
)

//go:embed head.html
var headTemplate string

//go:embed section.html
var sectionTemplate string

//go:embed sidebar.html
var sidebarTemplate string

//go:embed group.html
var groupTemplate string

func MakeHead(index string, noFTS bool) string {
	tmpl, err := template.New("header").Parse(headTemplate)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		CSS       string
		Index     string
		FTSSource string
		NoFTS     bool
	}{static.CSS, index, static.FTSExecutor, noFTS})
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func makeTOC(tree []fs.Entity) string {
	var list []string

	for _, i := range tree {
		switch v := i.(type) {
		case *fs.Group:
			list = append(list, fmt.Sprintf(`<details id="anchor-%s"><summary>%s</summary><div class="sidebar-group-wrapper">%s</div></details>`, v.CompoundID(), v.Meta().Title, makeTOC(v.Children)))
		case *fs.Section:
			compID := v.CompoundID()
			list = append(list, fmt.Sprintf(`<a href="#%s" id="anchor-%s">%s</a>`, compID, compID, v.Meta().Title))
		}
	}

	return strings.Join(list, "")
}

func MakeSidebar(tree []fs.Entity, version string, noFTS bool) string {
	TOC := makeTOC(tree)
	tmpl, err := template.New("sidebar").Parse(sidebarTemplate)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		TOC           string
		Version       string
		ThemeSelector string
		ToggleMenu 		string
		NoFTS         bool
	}{TOC, version, static.ThemeSelector, static.ToggleMenu, noFTS})
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func MakeContent(section *fs.Section) string {
	tmpl, err := template.New("content").Parse(sectionTemplate)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		HasTitle, HasSideNote        bool
		Title, SideNote, Content, ID string
	}{
		HasTitle:    section.Meta().Title != "",
		HasSideNote: section.HasSideNotes,
		ID:          section.CompoundID(),
		Title:       section.Meta().Title,
		SideNote:    markdown.ProcessSideNotes(section.SideNotes),
		Content:     markdown.ProcessContent(section.Content),
	})

	if err != nil {
		panic(err)
	}
	return buf.String()
}

func MakeGroup(group *fs.Group, fts *fts.FullTextSearchEngine) string {
	tmpl, err := template.New("content").Parse(groupTemplate)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer

	err = tmpl.Execute(&buf, struct {
		Title, ID, Child, Content string
	}{
		Title:   group.Meta().Title,
		ID:      group.CompoundID(),
		Content: RenderSectionNoMeta(group.Content),
		Child:   RenderItems(group.Children, fts),
	})

	if err != nil {
		panic(err)
	}
	return buf.String()
}

func RenderItems(items []fs.Entity, fts *fts.FullTextSearchEngine) string {
	var ret []string
	for _, e := range items {
		switch v := e.(type) {
		case *fs.Section:
			fts.AddSection(v)
			ret = append(ret, MakeContent(v))
		case *fs.Group:
			ret = append(ret, MakeGroup(v, fts))
		}
	}
	return strings.Join(ret, "")
}

func RenderSectionNoMeta(section *fs.Section) string {
	if section == nil {
		return ""
	}
	tmpl, err := template.New("content").Parse(sectionTemplate)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		HasTitle, HasSideNote        bool
		Title, SideNote, Content, ID string
	}{
		HasTitle:    false,
		HasSideNote: false,
		ID:          "",
		Title:       "",
		SideNote:    "",
		Content:     markdown.ProcessContent(section.Content),
	})

	if err != nil {
		panic(err)
	}
	return buf.String()
}
