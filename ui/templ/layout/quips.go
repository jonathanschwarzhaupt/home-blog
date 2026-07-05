package layout

import "math/rand/v2"

// footerQuips are shown one at a time, picked at random per render, in the
// site-wide footer — a small, deliberately subtle personality touch, not a
// call-to-action.
var footerQuips = []string{
	"Self-hosted in a homelab, which makes uptime a personal achievement, not a guarantee.",
	"No cookies were harmed in the making of this site. None were used, actually.",
	"Built with Go, Postgres, and an unreasonable amount of yak-shaving.",
	"If this page is down, it's probably my router, not Cloudflare.",
	"I wrote the metrics myself, so I know exactly how little is actually being tracked.",
	"This footer line rotates on every load. Yes, that was worth building instead of another post.",
	"Powered by caffeine, curiosity, and one very patient Postgres instance.",
}

func randomQuip() string {
	return footerQuips[rand.IntN(len(footerQuips))]
}
