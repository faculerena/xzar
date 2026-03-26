package model

import "time"

type ShortcutType string

const (
	ShortcutSubdomain ShortcutType = "subdomain"
	ShortcutPath      ShortcutType = "path"
)

type Shortcut struct {
	ID         int64
	Slug       string
	TargetURL  string
	Type       ShortcutType
	ClickCount int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type HomepageMode string

const (
	HomepageModeRedirect HomepageMode = "redirect"
	HomepageModeCarousel HomepageMode = "carousel"
)

type HomepageConfig struct {
	Mode        HomepageMode
	RedirectURL string
	UpdatedAt   time.Time
}

type CarouselImage struct {
	ID        int64
	Filename  string
	Original  string
	MimeType  string
	SortOrder int
	CreatedAt time.Time
}
