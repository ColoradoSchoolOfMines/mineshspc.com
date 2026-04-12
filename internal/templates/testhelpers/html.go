package testhelpers

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// CompareHTML parses both strings as HTML fragments and compares DOM trees,
// ignoring whitespace differences and normalizing attribute order.
func CompareHTML(t *testing.T, oldHTML, newHTML string) {
	t.Helper()
	bodyCtx := &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	}
	oldNodes, err := html.ParseFragment(strings.NewReader(oldHTML), bodyCtx)
	if err != nil {
		t.Fatalf("failed to parse old HTML: %v", err)
	}
	newNodes, err := html.ParseFragment(strings.NewReader(newHTML), bodyCtx)
	if err != nil {
		t.Fatalf("failed to parse new HTML: %v", err)
	}
	// Wrap in synthetic parent for comparison
	oldParent := &html.Node{Type: html.ElementNode, Data: "body"}
	for _, n := range oldNodes {
		oldParent.AppendChild(n)
	}
	newParent := &html.Node{Type: html.ElementNode, Data: "body"}
	for _, n := range newNodes {
		newParent.AppendChild(n)
	}
	if diff := diffChildren(oldParent, newParent); diff != "" {
		t.Errorf("HTML mismatch:\n%s", diff)
	}
}

func normalizeAttrVal(key, val string) string {
	switch key {
	case "class":
		// Normalize whitespace in class attribute values: trim edges and collapse internal spaces.
		return strings.Join(strings.Fields(val), " ")
	default:
		return val
	}
}

func normalizeAttrs(attrs []html.Attribute) []html.Attribute {
	out := make([]html.Attribute, 0, len(attrs))
	for _, a := range attrs {
		// Filter out spurious attributes that can arise from HTML syntax errors
		// e.g. a trailing comma in href="url", causes "," to be parsed as an attribute.
		key := strings.TrimSpace(a.Key)
		if key == "" || key == "," {
			continue
		}
		out = append(out, html.Attribute{
			Namespace: a.Namespace,
			Key:       key,
			Val:       normalizeAttrVal(key, a.Val),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func isSkippable(n *html.Node) bool {
	if n.Type == html.TextNode && strings.TrimSpace(n.Data) == "" {
		return true
	}
	if n.Type == html.CommentNode {
		return true
	}
	return false
}

func skipWhitespace(n *html.Node) *html.Node {
	for n != nil && isSkippable(n) {
		n = n.NextSibling
	}
	return n
}

func diffChildren(a, b *html.Node) string {
	ac := skipWhitespace(a.FirstChild)
	bc := skipWhitespace(b.FirstChild)
	for ac != nil || bc != nil {
		if diff := diffNodes(ac, bc); diff != "" {
			return diff
		}
		if ac != nil {
			ac = skipWhitespace(ac.NextSibling)
		}
		if bc != nil {
			bc = skipWhitespace(bc.NextSibling)
		}
	}
	return ""
}

func diffNodes(a, b *html.Node) string {
	if a == nil && b == nil {
		return ""
	}
	if a == nil {
		return fmt.Sprintf("extra node in new: %q (%v)", b.Data, b.Type)
	}
	if b == nil {
		return fmt.Sprintf("extra node in old: %q (%v)", a.Data, a.Type)
	}
	if a.Type != b.Type {
		return fmt.Sprintf("node type mismatch: old=%v(%q) new=%v(%q)", a.Type, a.Data, b.Type, b.Data)
	}
	if a.Type == html.ElementNode {
		if a.Data != b.Data {
			return fmt.Sprintf("element mismatch: old=<%s> new=<%s>", a.Data, b.Data)
		}
		oa, ob := normalizeAttrs(a.Attr), normalizeAttrs(b.Attr)
		if fmt.Sprint(oa) != fmt.Sprint(ob) {
			return fmt.Sprintf("attr mismatch on <%s>:\n  old=%v\n  new=%v", a.Data, oa, ob)
		}
		if diff := diffChildren(a, b); diff != "" {
			return fmt.Sprintf("in <%s>: %s", a.Data, diff)
		}
	}
	if a.Type == html.TextNode {
		at := strings.Join(strings.Fields(a.Data), " ")
		bt := strings.Join(strings.Fields(b.Data), " ")
		if at != bt {
			return fmt.Sprintf("text mismatch:\n  old=%q\n  new=%q", at, bt)
		}
	}
	return ""
}
