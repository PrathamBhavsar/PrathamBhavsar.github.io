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

type ProfileLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type Entry struct {
	Org  string `json:"org"`
	Role string `json:"role"`
	When string `json:"when"`
	URL  string `json:"url"`
	Desc string `json:"desc"`
}

type Profile struct {
	Name       string        `json:"name"`
	Tagline    string        `json:"tagline"`
	Location   string        `json:"location"`
	Email      string        `json:"email"`
	Links      []ProfileLink `json:"links"`
	About      []string      `json:"about"`
	Skills     []string      `json:"skills"`
	Experience []Entry       `json:"experience"`
	Education  []Entry       `json:"education"`
}

type Site struct {
	Projects []Project
	Profile  Profile
}

func loadSite() (*Site, error) {
	s := &Site{}
	raw, err := dataFS.ReadFile("data/projects.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &s.Projects); err != nil {
		return nil, err
	}
	for i := range s.Projects {
		if b, err := dataFS.ReadFile("data/content/" + s.Projects[i].Slug + ".md"); err == nil {
			s.Projects[i].Content = string(b)
		}
	}
	raw, err = dataFS.ReadFile("data/profile.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &s.Profile); err != nil {
		return nil, err
	}
	return s, nil
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
