package layout

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestBase_RendersOneOfTheKnownQuipsInFooter(t *testing.T) {
	var buf bytes.Buffer

	if err := Base("Test", "").Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}

	html := buf.String()

	found := false
	for _, q := range footerQuips {
		if strings.Contains(html, q) {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("rendered page doesn't contain any known footer quip")
	}
}
