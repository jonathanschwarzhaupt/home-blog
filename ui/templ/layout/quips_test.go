package layout

import "testing"

func TestRandomQuip_ReturnsOneOfTheKnownQuips(t *testing.T) {
	got := randomQuip()

	for _, q := range footerQuips {
		if q == got {
			return
		}
	}
	t.Fatalf("got quip %q, not present in footerQuips", got)
}

func TestRandomQuip_CanReturnDifferentValues(t *testing.T) {
	if len(footerQuips) < 2 {
		t.Skip("need at least 2 quips for this to mean anything")
	}

	seen := make(map[string]bool)
	for range 200 {
		seen[randomQuip()] = true
		if len(seen) > 1 {
			return
		}
	}
	t.Fatal("randomQuip returned the same value 200 times in a row — suspicious for a random pick")
}
