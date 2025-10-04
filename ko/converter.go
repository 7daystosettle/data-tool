package ko

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	kdl "github.com/sblinch/kdl-go"
	"github.com/sblinch/kdl-go/document"
)

const (
	commentNodeIdentifier = "_comment"
	textNodeIdentifier    = "_text"
)

// Ko holds a parsed document for conversion.
type Ko struct {
	doc *document.Document
}

// New parses XML from r into an internal KDL document model.
func NewFromXml(r io.Reader) (*Ko, error) {
	doc, err := xmlToKdl(r)
	if err != nil {
		return nil, fmt.Errorf("xmlToKdl: %w", err)
	}
	return &Ko{doc: doc}, nil
}

// NewFromKDL parses KDL from r into an internal document model.
func NewFromKdl(r io.Reader) (*Ko, error) {
	doc, err := kdl.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("kdl.Parse: %w", err)
	}
	return &Ko{doc: doc}, nil
}

// ToKdl writes a deterministic KDL representation to w.
func (e *Ko) ToKdl(w io.Writer) error {
	err := writeKDL(e.doc, w)
	if err != nil {
		return fmt.Errorf("writeKDL: %w", err)
	}
	return nil
}

func (e *Ko) ToXml(w io.Writer) error {
	var buf bytes.Buffer
	if err := kdlToXml(e.doc, &buf); err != nil {
		return fmt.Errorf("kdlToXml: %w", err)
	}
	out, err := selfCloseEmptyElements(buf.Bytes())
	if err != nil {
		return fmt.Errorf("selfCloseEmptyElements: %w", err)
	}
	if _, err := w.Write(out); err != nil {
		return fmt.Errorf("write out: %w", err)
	}
	return nil
}

func xmlToKdl(r io.Reader) (*document.Document, error) {
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch strings.ToLower(charset) {
		case "us-ascii":
			return input, nil
		default:
			return nil, fmt.Errorf("unsupported charset: %s", charset)
		}
	}
	doc := &document.Document{}
	var stack []*document.Node

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decoder.Token: %w", err)
		}

		var parent *document.Node
		if n := len(stack); n > 0 {
			parent = stack[n-1]
		}

		switch se := tok.(type) {
		case xml.StartElement:
			node := &document.Node{
				Name:       &document.Value{Value: se.Name.Local},
				Properties: make(document.Properties),
				Arguments:  []*document.Value{},
				Children:   []*document.Node{},
			}
			for _, a := range se.Attr {
				node.Properties[a.Name.Local] = &document.Value{Value: a.Value}
			}
			if parent != nil {
				parent.Children = append(parent.Children, node)
			} else {
				doc.Nodes = append(doc.Nodes, node)
			}
			stack = append(stack, node)

		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

		case xml.CharData:
			chunk := string(se)
			lines := strings.Split(chunk, "\n")
			for _, ln := range lines {
				t := strings.TrimSpace(ln)
				if t == "" {
					continue
				}
				if strings.HasPrefix(t, "//") {
					n := &document.Node{
						Name:       &document.Value{Value: commentNodeIdentifier},
						Properties: make(document.Properties),
						Arguments:  []*document.Value{{Value: strings.TrimSpace(strings.TrimPrefix(t, "//"))}},
						Children:   []*document.Node{},
					}
					if parent != nil {
						parent.Children = append(parent.Children, n)
					} else {
						doc.Nodes = append(doc.Nodes, n)
					}
					continue
				}
				if strings.HasPrefix(t, "<") || strings.HasPrefix(t, "property ") {
					fr := t
					if strings.HasPrefix(t, "property ") {
						fr = "<" + t + "/>"
					}
					nodes, perr := parseXMLFragment(fr)
					if perr == nil && len(nodes) > 0 {
						for _, n := range nodes {
							if parent != nil {
								parent.Children = append(parent.Children, n)
							} else {
								doc.Nodes = append(doc.Nodes, n)
							}
						}
						continue
					}
				}
				n := &document.Node{
					Name:       &document.Value{Value: textNodeIdentifier},
					Properties: make(document.Properties),
					Arguments:  []*document.Value{{Value: t}},
					Children:   []*document.Node{},
				}
				if parent != nil {
					parent.Children = append(parent.Children, n)
				} else {
					doc.Nodes = append(doc.Nodes, n)
				}
			}

		case xml.Comment:
			n := &document.Node{
				Name:       &document.Value{Value: commentNodeIdentifier},
				Properties: make(document.Properties),
				Arguments:  []*document.Value{{Value: string(se)}},
				Children:   []*document.Node{},
			}
			if parent != nil {
				parent.Children = append(parent.Children, n)
			} else {
				doc.Nodes = append(doc.Nodes, n)
			}
		}
	}
	return doc, nil
}

func parseXMLFragment(s string) ([]*document.Node, error) {
	dec := xml.NewDecoder(strings.NewReader("<frag>" + s + "</frag>"))
	dec.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch strings.ToLower(charset) {
		case "us-ascii":
			return input, nil
		default:
			return nil, fmt.Errorf("unsupported charset: %s", charset)
		}
	}
	var stack []*document.Node
	var out []*document.Node

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("fragment decoder.Token: %w", err)
		}

		switch se := tok.(type) {
		case xml.StartElement:
			if se.Name.Local == "frag" {
				continue
			}
			n := &document.Node{
				Name:       &document.Value{Value: se.Name.Local},
				Properties: make(document.Properties),
				Arguments:  []*document.Value{},
				Children:   []*document.Node{},
			}
			for _, a := range se.Attr {
				n.Properties[a.Name.Local] = &document.Value{Value: a.Value}
			}
			if len(stack) > 0 {
				stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, n)
			} else {
				out = append(out, n)
			}
			stack = append(stack, n)

		case xml.EndElement:
			if se.Name.Local == "frag" {
				continue
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

		case xml.CharData:
			txt := strings.TrimSpace(string(se))
			if txt != "" {
				n := &document.Node{
					Name:       &document.Value{Value: textNodeIdentifier},
					Properties: make(document.Properties),
					Arguments:  []*document.Value{{Value: txt}},
					Children:   []*document.Node{},
				}
				if len(stack) > 0 {
					stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, n)
				} else {
					out = append(out, n)
				}
			}
		}
	}
	return out, nil
}

func kdlToXml(doc *document.Document, w io.Writer) error {
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")

	_, err := w.Write([]byte(xml.Header))
	if err != nil {
		return fmt.Errorf("write xml header: %w", err)
	}

	err = kdlNodesToXml(doc.Nodes, enc)
	if err != nil {
		return fmt.Errorf("kdlNodesToXml: %w", err)
	}

	err = enc.Flush()
	if err != nil {
		return fmt.Errorf("encoder.Flush: %w", err)
	}
	return nil
}

func kdlNodesToXml(nodes []*document.Node, enc *xml.Encoder) error {
	for _, node := range nodes {
		switch node.Name.NodeNameString() {
		case commentNodeIdentifier:
			/* 			if len(node.Arguments) > 0 {
				err := enc.EncodeToken(xml.Comment(node.Arguments[0].ValueString()))
				if err != nil {
					return fmt.Errorf("encode comment: %w", err)
				}
			} */
			continue

		case textNodeIdentifier:
			if len(node.Arguments) > 0 {
				err := enc.EncodeToken(xml.CharData(node.Arguments[0].ValueString()))
				if err != nil {
					return fmt.Errorf("encode text: %w", err)
				}
			}
			continue
		}

		attrs := make([]xml.Attr, 0, len(node.Properties))

		orders := []string{"name", "trigger", "progression_name", "action", "cvar", "operation", "level", "value", "param1", "tier", "tags", "match_all_tags", "part", "active", "prefab", "parentTransform", "localPos"}
		orderedKeys := make(map[string]struct{})
		for _, key := range orders {
			if v, ok := node.Properties[key]; ok {
				attrs = append(attrs, xml.Attr{Name: xml.Name{Local: key}, Value: v.ValueString()})
				orderedKeys[key] = struct{}{}
			}
		}

		var rest []string
		for k := range node.Properties {
			if _, isOrdered := orderedKeys[k]; isOrdered {
				continue
			}
			rest = append(rest, k)
		}
		sort.Strings(rest)
		for _, k := range rest {
			attrs = append(attrs, xml.Attr{Name: xml.Name{Local: k}, Value: node.Properties[k].ValueString()})
		}

		start := xml.StartElement{Name: xml.Name{Local: node.Name.NodeNameString()}, Attr: attrs}
		err := enc.EncodeToken(start)
		if err != nil {
			return fmt.Errorf("encode start %q: %w", node.Name.NodeNameString(), err)
		}

		for _, a := range node.Arguments {
			err = enc.EncodeToken(xml.CharData(a.ValueString()))
			if err != nil {
				return fmt.Errorf("encode char data for %q: %w", node.Name.NodeNameString(), err)
			}
		}

		err = kdlNodesToXml(node.Children, enc)
		if err != nil {
			return fmt.Errorf("encode children for %q: %w", node.Name.NodeNameString(), err)
		}

		err = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: node.Name.NodeNameString()}})
		if err != nil {
			return fmt.Errorf("encode end %q: %w", node.Name.NodeNameString(), err)
		}
	}
	return nil
}

func writeKDL(doc *document.Document, w io.Writer) error {
	bw := bufio.NewWriter(w)
	for i, n := range doc.Nodes {
		err := emitNode(bw, n, 0)
		if err != nil {
			return fmt.Errorf("emitNode: %w", err)
		}
		if i < len(doc.Nodes)-1 {
			_, err = bw.WriteString("\n")
			if err != nil {
				return fmt.Errorf("write newline between top-level nodes: %w", err)
			}
		}
	}
	err := bw.Flush()
	if err != nil {
		return fmt.Errorf("bufio.Flush: %w", err)
	}
	return nil
}

func emitNode(w *bufio.Writer, n *document.Node, depth int) error {
	var err error
	name := n.Name.NodeNameString()

	if name == commentNodeIdentifier {
		if len(n.Arguments) == 0 {
			return nil
		}
		err := writeCommentLines(w, depth, n.Arguments[0].ValueString())
		if err != nil {
			return fmt.Errorf("writeCommentLines: %w", err)
		}
		return nil
	}

	// Special case: if a node has only one child and it's a _text node,
	// treat the text as an argument of the parent node.
	isInlineText := len(n.Children) == 1 &&
		n.Children[0].Name.NodeNameString() == textNodeIdentifier &&
		len(n.Children[0].Arguments) > 0

	if name == textNodeIdentifier && !isInlineText {
		indent(w, depth)

		err = writeKDLString(w, name)
		if err != nil {
			return fmt.Errorf("write text node name: %w", err)
		}
		_, err = w.WriteString(" ")
		if err != nil {
			return fmt.Errorf("write text node space: %w", err)
		}

		if len(n.Arguments) > 0 {
			err = writeKDLString(w, n.Arguments[0].ValueString())
			if err != nil {
				return fmt.Errorf("write text node value: %w", err)
			}
		}

		_, err = w.WriteString("\n")
		if err != nil {
			return fmt.Errorf("write text node newline: %w", err)
		}
		return nil
	}

	indent(w, depth)

	if strings.HasPrefix(name, "_") {
		err := writeKDLString(w, name)
		if err != nil {
			return fmt.Errorf("write quoted node name: %w", err)
		}
	} else {
		_, err := w.WriteString(name)
		if err != nil {
			return fmt.Errorf("write node name: %w", err)
		}
	}

	for _, a := range n.Arguments {
		_, err := w.WriteString(" ")
		if err != nil {
			return fmt.Errorf("write arg space: %w", err)
		}
		err = writeKDLString(w, a.ValueString())
		if err != nil {
			return fmt.Errorf("write arg value: %w", err)
		}
	}

	var keys []string
	prior := []string{"name", "trigger", "progression_name", "action", "cvar", "operation", "level", "value", "param1", "tier", "tags", "match_all_tags", "part", "active", "prefab", "parentTransform", "localPos"}
	inPrior := map[string]bool{}
	for _, k := range prior {
		if _, ok := n.Properties[k]; ok {
			keys = append(keys, k)
			inPrior[k] = true
		}
	}
	var rest []string
	for k := range n.Properties {
		if !inPrior[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	keys = append(keys, rest...)

	for _, k := range keys {
		_, err = w.WriteString(" " + k + "=")
		if err != nil {
			return fmt.Errorf("write prop key: %w", err)
		}
		err = writeKDLString(w, n.Properties[k].ValueString())
		if err != nil {
			return fmt.Errorf("write prop value: %w", err)
		}
	}

	if isInlineText {
		_, err = w.WriteString(" ")
		if err != nil {
			return fmt.Errorf("write inline text space: %w", err)
		}
		err = writeKDLString(w, n.Children[0].Arguments[0].ValueString())
		if err != nil {
			return fmt.Errorf("write inline text value: %w", err)
		}
	}

	if len(n.Children) == 0 || isInlineText {
		_, err = w.WriteString("\n")
		if err != nil {
			return fmt.Errorf("write node closing newline: %w", err)
		}
		return nil
	}

	_, err = w.WriteString(" {\n")
	if err != nil {
		return fmt.Errorf("write open brace: %w", err)
	}

	for _, c := range n.Children {
		if isInlineText && c.Name.NodeNameString() == textNodeIdentifier {
			continue
		}
		err = emitNode(w, c, depth+1)
		if err != nil {
			return fmt.Errorf("emit child: %w", err)
		}
	}

	indent(w, depth)

	_, err = w.WriteString("}\n")
	if err != nil {
		return fmt.Errorf("write close brace: %w", err)
	}
	return nil
}

func indent(w *bufio.Writer, depth int) {
	for i := 0; i < depth; i++ {
		_ = w.WriteByte(' ')
		_ = w.WriteByte(' ')
	}
}

func writeKDLString(w *bufio.Writer, s string) error {
	_, err := w.WriteString(`"` + escapeKDL(s) + `"`)
	if err != nil {
		return fmt.Errorf("write quoted string: %w", err)
	}
	return nil
}

func escapeKDL(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				_, _ = fmt.Fprintf(&b, `\u%04X`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func writeCommentLines(w *bufio.Writer, depth int, s string) error {
	// Normalize newlines
	txt := strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(txt, "\n")
	for i, ln := range lines {
		indent(w, depth)
		// Preserve empty lines as bare comment markers
		content := strings.TrimRight(ln, " \t")
		var toWrite string
		if content == "" {
			toWrite = "//\n"
		} else {
			toWrite = "// " + content + "\n"
		}
		_, err := w.WriteString(toWrite)
		if err != nil {
			return fmt.Errorf("write comment line %d: %w", i, err)
		}
	}
	// Ensure a trailing newline separation from following nodes when the
	// comment block is used where a node would be.
	// Callers already write their own newlines, so nothing extra here.
	return nil
}

var emptyElemRE = regexp.MustCompile(`(?s)<([A-Za-z_:][\w\.\-:]*)\b([^>]*)>\s*</[A-Za-z_:][\w\.\-:]*>`)

func selfCloseEmptyElements(in []byte) ([]byte, error) {
	// The regex finds elements with no content between the start and end tags,
	// allowing for whitespace. It captures the tag name and any attributes.
	// Then, it replaces the entire <tag...></tag> with a self-closing <tag... />.
	return emptyElemRE.ReplaceAllFunc(in, func(match []byte) []byte {
		submatches := emptyElemRE.FindSubmatch(match)
		if len(submatches) < 3 {
			return match
		}
		return []byte("<" + string(submatches[1]) + string(submatches[2]) + "/>")
	}), nil
}
