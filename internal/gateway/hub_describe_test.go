package gateway

import (
	"fmt"
	"strings"
	"testing"
)

// The instructions keep the whole tail and every server and tool name,
// even when the names alone exceed the usual description budget.
func TestHubAboutKeepsNames(t *testing.T) {
	const tail = "The tools of every MCP server in magpie's Library, each named <server>__<tool>."
	for _, n := range []int{1, 3, 12, 40} {
		for _, tools := range []int{0, 2, 30, 200} {
			ss := make([]HubServer, n)
			for i := range ss {
				ss[i] = HubServer{Name: fmt.Sprintf("srv%02d", i), About: strings.Repeat("Reads and writes things for you. ", 6), Known: true}
				for j := range tools {
					ss[i].Tools = append(ss[i].Tools, HubToolNote{Name: fmt.Sprintf("tool_%03d", j), About: strings.Repeat("word ", 60)})
				}
			}
			want := make(map[string]bool, n)
			byServer := make(map[string][]HubToolNote, n)
			for _, server := range ss {
				want[server.Name] = true
				byServer[server.Name] = server.Tools
			}
			say := hubAbout(ss, "This router: Los Angeles.", tail)
			if l := runeLen(say); tools == 0 && l > hubSayMax {
				t.Errorf("%d servers × %d tools: %d characters", n, tools, l)
			}
			if !strings.HasSuffix(say, tail) {
				t.Errorf("%d servers × %d tools: the tail is cut:\n%s", n, tools, say)
			}
			for _, line := range strings.Split(say, "\n")[1:] {
				if !strings.HasPrefix(line, "- ") {
					continue
				}
				server, _, _ := strings.Cut(strings.TrimPrefix(line, "- "), ": ")
				delete(want, server)
				for _, tool := range byServer[server] {
					if !strings.Contains(line, tool.Name) {
						t.Errorf("%s: missing tool %q", server, tool.Name)
						break
					}
				}
			}
			if len(want) != 0 {
				t.Errorf("%d servers × %d tools: missing detail lines for %v", n, tools, want)
			}
		}
	}
}

// A server's line names its tools with the first clause of each; a
// router's "[name] " before a description is left out, and a server
// magpie knows nothing of has no line.
func TestHubServerLine(t *testing.T) {
	s := HubServer{Name: "la", Title: "Magpie MCP hub · Los Angeles", Tools: []HubToolNote{
		{Name: "fetch__fetch", About: "[Los Angeles] Fetches a URL, and more."},
		{Name: "x", About: ""},
	}}
	if got, want := hubServerLine(s, 1000), "- la (Magpie MCP hub · Los Angeles): Tools: fetch__fetch (Fetches a URL), x."; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := hubServerLine(HubServer{Name: "bare"}, 1000); got != "" {
		t.Errorf("a server known by name only has a line: %q", got)
	}
	if got, want := hubServerLine(s, 10), "- la: Tools: fetch__fetch, x."; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Another magpie's first line keeps each server's words in brackets, so
// its clause isn't cut inside them.
func TestHubClauseBrackets(t *testing.T) {
	for in, want := range map[string]string{
		"Tools of fetch (Fetch a page, as Markdown), time.": "Tools of fetch (Fetch a page, as Markdown)",
		"Search [docs: all of them]. Then more.":            "Search [docs: all of them]",
		"Plain words, then more":                            "Plain words",
	} {
		if got := hubClause(in); got != want {
			t.Errorf("hubClause(%q) = %q, want %q", in, got, want)
		}
	}
}

// Another magpie's router says "Tools of a (…), b." of itself; the first
// line names it with what follows, not "Tools of" twice.
func TestHubLineOfARouter(t *testing.T) {
	got := HubLine([]HubServer{{Name: "la", About: "Tools of notes (Notes the team keeps), time."}})
	if want := "Tools of la (notes (Notes the team keeps))."; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// The first line is also written to Pi's configuration. Descriptions can
// give way to names, but the Magpie side never replaces names with a count.
func TestHubLineKeepsNames(t *testing.T) {
	names := strings.Fields("calc cloudflare-docs context7 cyberapi goanyapi ip2location layerfs magpie-memory filesystem github docker kubernetes postgres redis slack fetch time linear notion browser deepwiki sqlite playwright sentry grafana prometheus magpie-image")
	for _, n := range []int{10, 20, 27} {
		var ss []HubServer
		for _, name := range names[:n] {
			ss = append(ss, HubServer{Name: name, About: "Does some useful work for the calling agent."})
		}
		got := HubLine(ss)
		if strings.Contains(got, " more") {
			t.Errorf("%d servers: roster was shortened: %s", n, got)
		}
		parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(got, "Tools of "), "."), ", ")
		for i, name := range names[:n] {
			if i >= len(parts) || parts[i] != name && !strings.HasPrefix(parts[i], name+" (") {
				t.Errorf("%d servers: expected %q at position %d in %q", n, name, i, got)
				break
			}
		}
	}
}
