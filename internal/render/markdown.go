package render

import (
	"bytes"
	"html"
	"strconv"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

type headingInfo struct {
	Level int
	Text  string
	ID    string
}

// MarkdownToHTML converts Markdown to HTML and returns a Table of Contents HTML snippet.
func MarkdownToHTML(md string) (string, string, error) {
	var buf bytes.Buffer
	source := []byte(md)

	mdParser := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Linkify,
			extension.Strikethrough,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			gmhtml.WithUnsafe(),
		),
	)

	reader := text.NewReader(source)
	doc := mdParser.Parser().Parse(reader)

	toc := buildTOCHTML(doc, source)

	if err := mdParser.Renderer().Render(&buf, source, doc); err != nil {
		return "", "", err
	}

	return buf.String(), toc, nil
}

func buildTOCHTML(doc ast.Node, source []byte) string {
	headings := collectHeadings(doc, source)
	if len(headings) == 0 {
		return ""
	}

	baseLevel := headings[0].Level
	for _, h := range headings {
		if h.Level < baseLevel {
			baseLevel = h.Level
		}
	}

	var b strings.Builder
	b.WriteString(`<nav class="toc-nav" aria-label="Table of contents">`)

	currentDepth := 0
	openItem := false

	for _, h := range headings {
		depth := h.Level - baseLevel + 1
		if depth < 1 {
			depth = 1
		}

		for currentDepth > depth {
			if openItem {
				b.WriteString("</li>")
				openItem = false
			}
			b.WriteString("</ol>")
			currentDepth--
		}

		for currentDepth < depth {
			listType := listTypeForDepth(currentDepth + 1)
			b.WriteString(`<ol class="toc-list toc-depth-`)
			b.WriteString(strconv.Itoa(currentDepth + 1))
			b.WriteString(`" type="`)
			b.WriteString(listType)
			b.WriteString(`">`)
			currentDepth++
		}

		if openItem {
			b.WriteString("</li>")
		}

		b.WriteString(`<li><a href="#`)
		b.WriteString(h.ID)
		b.WriteString(`">`)
		b.WriteString(html.EscapeString(h.Text))
		b.WriteString(`</a>`)

		openItem = true
	}

	if openItem {
		b.WriteString("</li>")
	}

	for currentDepth > 0 {
		b.WriteString("</ol>")
		currentDepth--
	}

	b.WriteString("</nav>")
	return b.String()
}

func listTypeForDepth(depth int) string {
	switch depth {
	case 1:
		return "1"
	case 2:
		return "a"
	case 3:
		return "i"
	default:
		return "1"
	}
}

func collectHeadings(doc ast.Node, source []byte) []headingInfo {
	var headings []headingInfo
	seen := make(map[string]int)

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		if h.Level < 2 || h.Level > 4 {
			return ast.WalkContinue, nil
		}

		text := strings.TrimSpace(string(h.Text(source)))
		if text == "" {
			return ast.WalkContinue, nil
		}

		var id string
		if attr, ok := h.AttributeString("id"); ok {
			if s, ok := attr.(string); ok {
				id = s
			} else if b, ok := attr.([]byte); ok {
				id = string(b)
			}
		}
		if id == "" {
			id = slugify(text)
		}
		id = ensureUniqueID(id, seen)
		h.SetAttributeString("id", id)

		headings = append(headings, headingInfo{
			Level: h.Level,
			Text:  text,
			ID:    id,
		})

		return ast.WalkContinue, nil
	})

	return headings
}

func slugify(input string) string {
	input = strings.ToLower(input)
	var b strings.Builder
	lastHyphen := false

	for _, r := range input {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastHyphen = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		case unicode.IsSpace(r) || r == '-':
			if !lastHyphen && b.Len() > 0 {
				b.WriteRune('-')
				lastHyphen = true
			}
		default:
			// Skip other characters
		}
	}

	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "section"
	}
	return slug
}

func ensureUniqueID(base string, seen map[string]int) string {
	if base == "" {
		base = "section"
	}

	if count, exists := seen[base]; exists {
		count++
		seen[base] = count
		return base + "-" + strconv.Itoa(count)
	}

	seen[base] = 0
	return base
}
