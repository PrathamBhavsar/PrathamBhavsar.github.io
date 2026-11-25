package main

import (
	"embed"
	"encoding/json"
	"regexp"
	"strings"
)

//go:embed data
var dataFS embed.FS

type Link struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
	Href  string `json:"href"`
}

type Project struct {
	Slug    string   `json:"slug"`
	Title   string   `json:"title"`
	Line    string   `json:"one_liner"`
	Role    string   `json:"role"`
	Spec    []string `json:"spec"`
	Stack   []string `json:"stack"`
	Note    string   `json:"note"`
	Links   []Link   `json:"links"`
	Content string   `json:"-"`
}

type Site struct{ Projects []Project }

func loadSite() (*Site, error) {
	raw, err := dataFS.ReadFile("data/projects.json")
	if err != nil {
		return nil, err
	}
	var ps []Project
	if err := json.Unmarshal(raw, &ps); err != nil {
		return nil, err
	}
	for i := range ps {
		if b, err := dataFS.ReadFile("data/content/" + ps[i].Slug + ".md"); err == nil {
			ps[i].Content = string(b)
		}
	}
	return &Site{Projects: ps}, nil
}

var (
	reCode = regexp.MustCompile("`([^`]+)`")
	reBold = regexp.MustCompile(`\*\*([^*]+)\*\*`)
)

// strip turns the tiny markdown subset the case studies use into plain text.
func strip(s string) string {
	s = reCode.ReplaceAllString(s, "$1")
	s = reBold.ReplaceAllString(s, "$1")
	return strings.TrimSpace(s)
}
