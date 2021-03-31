package markdown

import (
	"fmt"
	"strings"

	"github.com/heyvito/docuowl/markdown/ast"
	"github.com/heyvito/docuowl/markdown/html"
	"github.com/heyvito/docuowl/markdown/parser"
)

func convertToAST(input []string) (ast.Node, error) {
	buf := []byte(strings.Join(input, "\n"))
	node := Parse(buf, parser.New())

	if _, ok := node.(*ast.Document); !ok {
		return nil, fmt.Errorf("parser did not return a Document object")
	}

	return node, nil
}

func convertToHTML(node ast.Node) string {
	renderer := html.NewRenderer(html.RendererOptions{
		Flags: html.CommonFlags,
	})

	return string(Render(node, renderer))
}

func isList(node ast.Node) bool {
	_, ok := node.(*ast.List)
	return ok
}

func isParagraph(node ast.Node) bool {
	_, ok := node.(*ast.Paragraph)
	return ok
}

func unlinkNode(n ast.Node) {
	if p := n.GetParent(); p != nil {
		newChildren := make([]ast.Node, 0, len(p.GetChildren()))
		for _, nn := range p.GetChildren() {
			if nn != n {
				newChildren = append(newChildren, nn)
			}
		}
		p.SetChildren(newChildren)
	}
	n.SetParent(nil)
}

func stringifyParagraph(node *ast.Paragraph) string {
	var text []string
	for _, c := range node.Children {
		switch v := c.(type) {
		case *ast.Text:
			text = append(text, string(v.Literal))
		default:
			fmt.Printf("WARNING: stringifyParagraph found unexpected type %t\n", c)
		}
	}
	return strings.Join(text, "")
}

//goland:noinspection ALL
func processListAsOwlElement(node *ast.List) []ast.Node {
	unlinkNode(node)
	// Node should be a list
	if !isList(node) {
		fmt.Printf("[WARNING] Non-list node reached processListAsOwlElement: %#v\n", node)
		return []ast.Node{}
	}

	var result []ast.Node
	var item *ast.DocuowlListItem
	for idx, child := range node.Children {
		unlinkNode(child)
		if idx%2 == 0 {
			if item != nil {
				result = append(result, item)
			}

			switch len(child.GetChildren()) {
			case 2:
				first := child.GetChildren()[0]
				second := child.GetChildren()[1]
				if isParagraph(first) && isList(second) {
					item := &ast.DocuowlList{Title: stringifyParagraph(first.(*ast.Paragraph)), HasTitle: true}
					unlinkNode(first)
					second.SetParent(item)
					item.Children = processListAsOwlElement(second.(*ast.List))
					result = append(result, item)
				} else {
					item = &ast.DocuowlListItem{Title: []ast.Node{first}}
					item.Children = []ast.Node{second}
				}
			case 1:
				first := child.GetChildren()[0]
				if isList(first) {
					item.Children = append(item.Children, processListAsOwlElement(first.(*ast.List))...)
					break
				}
				fallthrough
			default:
				item = &ast.DocuowlListItem{Title: child.GetChildren()}
			}
		} else {
			for _, c := range child.GetChildren() {
				unlinkNode(c)

				if isList(c) {
					processedData := processListAsOwlElement(c.(*ast.List))
					if len(processedData) == 1 && isOwlList(processedData[0]) {
						item.Children = append(item.Children, processedData[0])
					} else {
						otherList := &ast.DocuowlList{
							HasTitle: false,
						}
						otherList.Children = processedData
						item.Children = append(item.Children, otherList)
					}
				} else {
					item.Children = append(item.Children, c)
				}
				c.SetParent(item)
			}
		}
	}
	if item != nil {
		result = append(result, item)
	}

	return result
}

func isOwlList(node ast.Node) bool {
	_, ok := node.(*ast.DocuowlList)
	return ok
}

func shouldProcessChild(node ast.Node) bool {
	_, isBox := node.(*ast.DocuowlBox)
	_, isHR := node.(*ast.HorizontalRule)
	return !isBox && !isHR || isList(node)
}

func ProcessContent(input []string) string {
	root, err := convertToAST(input)
	if err != nil {
		panic(err)
	}
	isProcessingList := false
	var currentList *ast.DocuowlList
	ast.WalkFunc(root, func(node ast.Node, entering bool) ast.WalkStatus {
		if shouldProcessChild(node) && entering && isProcessingList {
			list, ok := node.(*ast.List)
			if !ok {
				// Not a list. Just stop processing it as a list, and continue.
				isProcessingList = false
				return ast.GoToNext
			}

			ast.AppendChildren(currentList, processListAsOwlElement(list)...)

			return ast.SkipChildren
		}

		if !entering {
			return ast.GoToNext
		}

		if header, ok := node.(*ast.Heading); ok {
			if header.Level < 3 {
				header.Level = 3
			}
		}

		if _, ok := node.(*ast.HorizontalRule); ok && isProcessingList {
			isProcessingList = false
			ast.RemoveFromTree(node)
			return ast.GoToNext
		}

		if l, ok := node.(*ast.DocuowlList); ok {
			currentList = l
			isProcessingList = true
		}

		return ast.GoToNext
	})
	return convertToHTML(root)
}

func ProcessSideNotes(input []string) string {
	root, err := convertToAST(input)
	if err != nil {
		panic(err)
	}

	isProcessingBox := false
	var currentBox *ast.DocuowlBox

	ast.WalkFunc(root, func(node ast.Node, entering bool) ast.WalkStatus {
		if shouldProcessChild(node) && entering && isProcessingBox {
			// ast.RemoveFromTree does weird things, and we have no time.
			unlinkNode(node)
			ast.AppendChild(currentBox, node)
			node.SetParent(currentBox)
			return ast.SkipChildren
		}

		if !entering {
			return ast.GoToNext
		}

		if h, ok := node.(*ast.Heading); ok {
			h.Level = 2
			return ast.GoToNext
		}

		if _, ok := node.(*ast.HorizontalRule); ok {
			if isProcessingBox {
				isProcessingBox = false
				ast.RemoveFromTree(node)
			}
			return ast.GoToNext
		}

		if b, ok := node.(*ast.DocuowlBox); ok {
			isProcessingBox = true
			currentBox = b
		}

		return ast.GoToNext
	})

	return convertToHTML(root)
}
