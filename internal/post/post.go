package post

import (
	"strings"
	"time"
)

// Post represents a single forum post.
type Post struct {
	ID          int       `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Username    string    `json:"username"`
	ContentRaw  string    `json:"raw"`
	ContentHTML string    `json:"cooked"`
	Title       string    `json:"topic_title"`
}

// ContainsKeywords checks if the post has least one of the keywords in the post title, and/or content.
func (p *Post) ContainsKeywords(keywords string) bool {
	if keywords == "" {
		return false
	}

	for _, keyword := range strings.Split(keywords, ",") {
		if strings.Contains(p.ContentRaw, keyword) || strings.Contains(p.Title, keyword) {
			return true
		}
	}

	return false
}
