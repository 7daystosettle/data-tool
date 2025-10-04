package ko

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/sblinch/kdl-go/document"
)

const sampleXml = `<?xml version="1.0" encoding="UTF-8"?>
<!-- sample comment -->
<root a="b">
	<child1 c="d"/>
	<child2>some text</child2>
</root>`

func TestXmlToKdl(t *testing.T) {
	doc, err := xmlToKdl(strings.NewReader(sampleXml))
	if err != nil {
		t.Fatalf("XmlToKdl failed: %v", err)
	}

	expectedDoc := &document.Document{}
	commentNode := &document.Node{Name: &document.Value{Value: "_comment"}, Properties: make(document.Properties), Arguments: []*document.Value{}, Children: []*document.Node{}}
	commentNode.Arguments = append(commentNode.Arguments, &document.Value{Value: " sample comment "})
	expectedDoc.Nodes = append(expectedDoc.Nodes, commentNode)

	rootNode := &document.Node{Name: &document.Value{Value: "root"}, Properties: make(document.Properties), Arguments: []*document.Value{}, Children: []*document.Node{}}
	rootNode.Properties["a"] = &document.Value{Value: "b"}
	expectedDoc.Nodes = append(expectedDoc.Nodes, rootNode)

	child1Node := &document.Node{Name: &document.Value{Value: "child1"}, Properties: make(document.Properties), Arguments: []*document.Value{}, Children: []*document.Node{}}
	child1Node.Properties["c"] = &document.Value{Value: "d"}
	rootNode.Children = append(rootNode.Children, child1Node)

	child2Node := &document.Node{Name: &document.Value{Value: "child2"}, Properties: make(document.Properties), Arguments: []*document.Value{}, Children: []*document.Node{}}
	textNode := &document.Node{Name: &document.Value{Value: "_text"}, Properties: make(document.Properties), Arguments: []*document.Value{}, Children: []*document.Node{}}
	textNode.Arguments = append(textNode.Arguments, &document.Value{Value: "some text"})
	child2Node.Children = append(child2Node.Children, textNode)
	rootNode.Children = append(rootNode.Children, child2Node)

	if !reflect.DeepEqual(doc, expectedDoc) {
		t.Errorf("XmlToKdl output mismatch:\nExpected:\n%v\nGot:\n%v", expectedDoc, doc)
	}
}

func TestKdlToXml(t *testing.T) {
	doc, err := xmlToKdl(strings.NewReader(sampleXml))
	if err != nil {
		t.Fatalf("XmlToKdl failed during setup: %v", err)
	}

	var buf bytes.Buffer
	if err := kdlToXml(doc, &buf); err != nil {
		t.Fatalf("KdlToXml failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "<!-- sample comment -->") {
		t.Errorf("KdlToXml output missing comment")
	}
	if !strings.Contains(got, "<root a=\"b\">") {
		t.Errorf("KdlToXml output missing root element or attribute")
	}
	if !strings.Contains(got, "<child1 c=\"d\">") {
		t.Errorf("KdlToXml output missing child1 element or attribute")
	}
	if !strings.Contains(got, ">some text</child2>") {
		t.Errorf("KdlToXml output missing text content")
	}
}
