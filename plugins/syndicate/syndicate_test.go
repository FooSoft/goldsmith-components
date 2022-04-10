package syndicate

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
		AtomPath:    "feed.atom",
		RssPath:     "feed.xml",
		JsonPath:    "feed.json",
	}

	feedConfig.ItemConfig = ItemConfig{
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
				Chain(New("https://foosoft.net", "FeedName").WithFeed("posts", feedConfig))
		},
	)
}
