package render

import (
    "bytes"
    "github.com/yuin/goldmark"
)

func MarkdownToHTML(md string) (string, error) {
    var buf bytes.Buffer

    // Create Markdown parser
    mdParser := goldmark.New()

    // Convert Markdown → HTML
    if err := mdParser.Convert([]byte(md), &buf); err != nil {
        return "", err
    }

    return buf.String(), nil
}


/** What & Why's:
    - goldmark package (industry standard)
*/