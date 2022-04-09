package rss

import (
	"bytes"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/FooSoft/goldsmith"
	"github.com/gorilla/feeds"
)

type item struct {
	title       string
	authorName  string
	authorEmail string
	description string
	id          string
	updated     time.Time
	created     time.Time
	content     string

	url string
}

type itemList []item

func (self itemList) Len() int {
	return len(self)
}

func (self itemList) Less(i, j int) bool {
	if less := self[i].created.Before(self[j].created); less {
		return true
	}

	if less := self[i].updated.Before(self[j].updated); less {
		return true
	}

	if less := strings.Compare(self[i].id, self[j].id) < 0; less {
		return true
	}

	if less := strings.Compare(self[i].title, self[j].title) < 0; less {
		return true
	}

	return false
}

func (self itemList) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type FeedConfig struct {
	Title       string
	Url         string
	Description string
	AuthorName  string
	AuthorEmail string
	Updated     time.Time
	Created     time.Time
	Id          string
	Subtitle    string
	Copyright   string
	ImageUrl    string
}

type ItemConfig struct {
	BaseUrl        string
	RssEnableKey   string
	TitleKey       string
	AuthorNameKey  string
	AuthorEmailKey string
	DescriptionKey string
	IdKey          string
	UpdatedKey     string
	CreatedKey     string
	ContentKey     string
}

// Rss chainable context.
type Rss struct {
	feedConfig FeedConfig
	itemConfig ItemConfig

	atomPath string
	jsonPath string
	rssPath  string

	items itemList
	lock  sync.Mutex
}

// New creates a new instance of the Rss plugin
func New(feedConfig FeedConfig, itemConfig ItemConfig) *Rss {
	return &Rss{
		feedConfig: feedConfig,
		itemConfig: itemConfig,
	}
}

func (*Rss) Name() string {
	return "rss"
}

func (self *Rss) AtomPath(path string) *Rss {
	self.atomPath = path
	return self
}

func (self *Rss) JsonPath(path string) *Rss {
	self.jsonPath = path
	return self
}

func (self *Rss) RssPath(path string) *Rss {
	self.rssPath = path
	return self
}

func (self *Rss) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	defer context.DispatchFile(inputFile)

	getString := func(key string) string {
		if len(key) == 0 {
			return ""
		}

		prop, ok := inputFile.Prop(key)
		if !ok {
			return ""
		}

		result, ok := prop.(string)
		if !ok {
			return ""
		}

		return result
	}

	getDate := func(key string) time.Time {
		if len(key) == 0 {
			return time.Time{}
		}

		prop, ok := inputFile.Prop(key)
		if !ok {
			return time.Time{}
		}

		result, ok := prop.(time.Time)
		if !ok {
			return time.Time{}
		}

		return result
	}

	getBool := func(key string) bool {
		if len(key) == 0 {
			return false
		}

		prop, ok := inputFile.Prop(key)
		if !ok {
			return false
		}

		result, ok := prop.(bool)
		if !ok {
			return false
		}

		return result
	}

	if rssEnable := getBool(self.itemConfig.RssEnableKey); !rssEnable {
		return nil
	}

	item := item{
		title:       getString(self.itemConfig.TitleKey),
		authorName:  getString(self.itemConfig.AuthorNameKey),
		authorEmail: getString(self.itemConfig.AuthorEmailKey),
		description: getString(self.itemConfig.DescriptionKey),
		id:          getString(self.itemConfig.IdKey),
		updated:     getDate(self.itemConfig.UpdatedKey),
		created:     getDate(self.itemConfig.CreatedKey),
		content:     getString(self.itemConfig.ContentKey),
		url:         path.Join(self.itemConfig.BaseUrl, inputFile.Path()),
	}

	self.lock.Lock()
	self.items = append(self.items, item)
	self.lock.Unlock()

	return nil
}

func (self *Rss) Finalize(context *goldsmith.Context) error {
	feed := feeds.Feed{
		Title:       self.feedConfig.Title,
		Link:        &feeds.Link{Href: self.feedConfig.Url},
		Description: self.feedConfig.Description,
		Updated:     self.feedConfig.Updated,
		Created:     self.feedConfig.Created,
		Id:          self.feedConfig.Id,
		Subtitle:    self.feedConfig.Subtitle,
		Copyright:   self.feedConfig.Copyright,
	}

	if len(self.feedConfig.AuthorName) > 0 || len(self.feedConfig.AuthorEmail) > 0 {
		feed.Author = &feeds.Author{
			Name:  self.feedConfig.AuthorName,
			Email: self.feedConfig.AuthorEmail,
		}
	}

	if len(self.feedConfig.ImageUrl) > 0 {
		feed.Image = &feeds.Image{Url: self.feedConfig.ImageUrl}
	}

	sort.Sort(self.items)

	for _, item := range self.items {
		feedItem := feeds.Item{
			Title:       item.title,
			Description: item.description,
			Id:          item.id,
			Updated:     item.updated,
			Created:     item.created,
			Content:     item.content,
			Link:        &feeds.Link{Href: item.url},
		}

		if len(item.authorName) > 0 || len(item.authorEmail) > 0 {
			feedItem.Author = &feeds.Author{
				Name:  item.authorName,
				Email: item.authorEmail,
			}
		}

		feed.Items = append(feed.Items, &feedItem)
	}

	if len(self.atomPath) > 0 {
		var buff bytes.Buffer
		if err := feed.WriteAtom(&buff); err != nil {
			return err
		}

		file, err := context.CreateFileFromReader(self.atomPath, bytes.NewReader(buff.Bytes()))
		if err != nil {
			return err
		}

		context.DispatchFile(file)
	}

	if len(self.rssPath) > 0 {
		var buff bytes.Buffer
		if err := feed.WriteRss(&buff); err != nil {
			return err
		}

		file, err := context.CreateFileFromReader(self.rssPath, bytes.NewReader(buff.Bytes()))
		if err != nil {
			return err
		}

		context.DispatchFile(file)
	}

	if len(self.jsonPath) > 0 {
		var buff bytes.Buffer
		if err := feed.WriteJSON(&buff); err != nil {
			return err
		}

		file, err := context.CreateFileFromReader(self.jsonPath, bytes.NewReader(buff.Bytes()))
		if err != nil {
			return err
		}

		context.DispatchFile(file)
	}

	return nil
}
