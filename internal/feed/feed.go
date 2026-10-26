// Package feed renders the private podcast feed: RSS 2.0 with the iTunes
// namespace that podcast apps expect.
package feed

import (
	"encoding/xml"
	"fmt"
	"io"
	"time"
)

type Channel struct {
	Title       string
	Description string
	Link        string // the site
	Items       []Item
}

type Item struct {
	GUID        string
	Title       string
	Description string
	Link        string // the episode page, once it exists
	AudioURL    string
	AudioBytes  int64
	Duration    time.Duration // zero: omitted
	Published   time.Time
}

type rss struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	ITunes  string     `xml:"xmlns:itunes,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	Explicit    string    `xml:"itunes:explicit"`
	Block       string    `xml:"itunes:block"` // keep it out of any directory that finds the URL
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string       `xml:"title"`
	Description string       `xml:"description"`
	Link        string       `xml:"link,omitempty"`
	GUID        rssGUID      `xml:"guid"`
	PubDate     string       `xml:"pubDate"`
	Enclosure   rssEnclosure `xml:"enclosure"`
	Duration    string       `xml:"itunes:duration,omitempty"`
}

type rssGUID struct {
	IsPermaLink string `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

type rssEnclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

// Write renders c as RSS.
func Write(w io.Writer, c Channel) error {
	out := rss{
		Version: "2.0",
		ITunes:  "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Channel: rssChannel{
			Title:       c.Title,
			Link:        c.Link,
			Description: c.Description,
			Language:    "en",
			Explicit:    "false",
			Block:       "Yes",
		},
	}
	for _, it := range c.Items {
		ri := rssItem{
			Title:       it.Title,
			Description: it.Description,
			Link:        it.Link,
			GUID:        rssGUID{IsPermaLink: "false", Value: it.GUID},
			PubDate:     it.Published.UTC().Format(time.RFC1123Z),
			Enclosure:   rssEnclosure{URL: it.AudioURL, Length: it.AudioBytes, Type: "audio/mpeg"},
		}
		if it.Duration > 0 {
			s := int(it.Duration.Round(time.Second).Seconds())
			ri.Duration = fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
		}
		out.Channel.Items = append(out.Channel.Items, ri)
	}
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(out); err != nil {
		return err
	}
	_, err := io.WriteString(w, "\n")
	return err
}
