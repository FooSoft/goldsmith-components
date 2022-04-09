package rss

import (
	"testing"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/harness"
	"github.com/FooSoft/goldsmith-components/plugins/frontmatter"
)

func Test(self *testing.T) {
	feedConfig := FeedConfig{
		Title:       "Feed Title",
		Url:         "https://foosoft.net",
		Description: "Feed Description",
		AuthorName:  "Author Name",
		AuthorEmail: "Author Email",
		Id:          "Feed Id",
		Subtitle:    "Feed Subtitle",
		Copyright:   "Feed Copyright",
		ImageUrl:    "Feed Image Url",
	}

	itemConfig := ItemConfig{
		BaseUrl:        "https://foosoft.net",
		RssEnableKey:   "RssEnable",
		TitleKey:       "Title",
		AuthorNameKey:  "AuthorName",
		AuthorEmailKey: "AuthorEmail",
		DescriptionKey: "Description",
		IdKey:          "Id",
		UpdatedKey:     "Updated",
		CreatedKey:     "Created",
		ContentKey:     "Content",
	}

	harness.Validate(
		self,
		func(gs *goldsmith.Goldsmith) {
			gs.
				Chain(frontmatter.New()).
				Chain(New(feedConfig, itemConfig).RssPath("feed.xml").AtomPath("feed.atom").JsonPath("feed.json"))
		},
	)
}
