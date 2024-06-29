package hugoembedding

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Parse is a function to parse markdown for chunks to convert
// to embeddings
// it takes a byte slice and returns a slice of Chunk pointers and an error.
func Parse(source []byte) (*[]Chunk, error) {
	md := goldmark.New(goldmark.WithExtensions())
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var chunks []Chunk = make([]Chunk, 0)

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			kind := n.Kind()
			Logger.Debug("Node", "kind", kind.String())
			switch kind {
			case ast.KindHeading:
				heading := n.(*ast.Heading)
				// we look for h1 or h2 headings
				if heading.Level == 1 || heading.Level == 2 {
					return ast.WalkContinue, nil
				}
			case ast.KindFencedCodeBlock:
				// n.Dump(source, 2)
				codeSpanText := extractFencedCodeBlocks(n, source)
				aChunk := Chunk{
					Chunk:     &codeSpanText,
					Context:   &codeSpanText,
					Reference: nil,
				}
				chunks = append(chunks, aChunk)
				Logger.Debug("Cunks from markdown", "chunk", *aChunk.Chunk)

				return ast.WalkContinue, nil
			case ast.KindParagraph:
				prev := n.PreviousSibling()
				if prev != nil {
					if prev.Kind() == ast.KindHeading {
						heading := prev.(*ast.Heading)
						if heading.Level == 1 || heading.Level == 2 {
							headingText := heading.Text(source)
							// Get text from paragraph
							// paragraphText := n.(*ast.Paragraph)

							paragraphText := string(headingText) +
								"\n" +
								extractTextFromParagraph(n.(*ast.Paragraph), source) +
								"\n"
							aChunk := Chunk{
								Chunk:     &paragraphText,
								Context:   &paragraphText,
								Reference: nil,
							}
							chunks = append(chunks, aChunk)
							Logger.Debug("Cunks", "chunk", *aChunk.Chunk)
						}
					} else {
						paragraphText := "\n" +
							extractTextFromParagraph(n.(*ast.Paragraph), source) +
							"\n"
						aChunk := Chunk{
							Chunk:     &paragraphText,
							Context:   &paragraphText,
							Reference: nil,
						}
						chunks = append(chunks, aChunk)
						Logger.Debug("Cunks", "chunk", *aChunk.Chunk)
					}

				}
			case ast.KindText:
				text := string(n.(*ast.Text).Text(source))
				aChunk := Chunk{
					Chunk:     &text,
					Context:   &text,
					Reference: nil,
				}
				chunks = append(chunks, aChunk)
				Logger.Debug("Cunks", "chunk", *aChunk.Chunk)

			case ast.KindList:
				text := extractTextFromList(n.(*ast.List), source)
				aChunk := Chunk{
					Chunk:     &text,
					Context:   &text,
					Reference: nil,
				}
				chunks = append(chunks, aChunk)
			}
		}
		return ast.WalkContinue, nil
	})

	return &chunks, nil
}

func extractTextFromParagraph(paragraph *ast.Paragraph, source []byte) string {
	var buffer bytes.Buffer

	// Iterate through the children of the paragraph node
	for child := paragraph.FirstChild(); child != nil; child = child.NextSibling() {
		switch child := child.(type) {
		case *ast.Text:
			buffer.Write(child.Text(source))
		case *ast.String:
			buffer.Write(child.Value)
		case *ast.CodeSpan:
			// we're only interested in code blocks
			codeSpanText := extractCodeSpanText(child, source)
			buffer.WriteString(codeSpanText)
		case *ast.Emphasis:
			// If there's emphasis, you have to iterate through the children of this node as well
			for subChild := child.FirstChild(); subChild != nil; subChild = subChild.NextSibling() {
				switch subChild := subChild.(type) {
				case *ast.Text:
					buffer.Write(subChild.Text(source))
				case *ast.String:
					buffer.Write(subChild.Value)
				}
			}
		case *ast.Link:
			// If it's a link, we extract the link text
			// ###TODO
			// buffer.WriteString(extractTextFromParagraph(child, source))
		}
	}
	// Add a space if it wasn't the last node and buffer isn't empty to simulate spaces that were implicitly between words/nodes
	for child := paragraph.FirstChild(); child != nil; child = child.NextSibling() {
		if buffer.Len() > 0 && child.NextSibling() != nil {
			buffer.WriteByte(' ')
		}
	}

	return buffer.String()
}

func extractCodeSpanText(node ast.Node, source []byte) string {
	var codeSpanText string
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n := n.(type) {
			case *ast.CodeSpan:
				codeSpanText += string(n.Text(source))
			}
		}
		return ast.WalkContinue, nil
	})
	return codeSpanText
}

func extractFencedCodeBlocks(node ast.Node, source []byte) string {
	var codeBlockContent string
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if fc, ok := n.(*ast.FencedCodeBlock); ok {
				// Read from the beginning line to the ending line (exclusive) of the FencedCodeBlock content lines
				lines := fc.Lines()
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					codeBlockContent += string(line.Value(source))
				}
			}
		}
		return ast.WalkContinue, nil
	})
	return codeBlockContent
}

func extractTextFromList(node ast.Node, source []byte) string {
	var buf bytes.Buffer
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			kind := n.Kind()
			Logger.Debug("Node", "kind", kind.String())
			switch t := n.(type) {
			case *ast.ListItem:
				buf.Write([]byte(" - "))
			case *ast.Text:
				buf.Write(t.Segment.Value(source))
			case *ast.String:
				buf.WriteString(string(t.Value))
			default:
				return ast.WalkContinue, nil
			}
		}
		return ast.WalkContinue, nil
	})
	return buf.String()
}
