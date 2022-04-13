package postchecker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/halkyon/discourse-scanner/internal/post"
)

const requestTimeoutSeconds = 10

// Posts represents a response containing a number of posts.
type Posts struct {
	Latest []post.Post `json:"latest_posts"`
}

// PostChecker represents a way of checking forum posts for keywords.
type PostChecker struct {
	client   http.Client
	baseURL  string
	keywords string
	interval time.Duration
}

// New returns a new instance of PostChecker.
func New(baseURL, keywords string, interval time.Duration) *PostChecker {
	return &PostChecker{
		client:   http.Client{Timeout: requestTimeoutSeconds * time.Second},
		baseURL:  baseURL,
		keywords: keywords,
		interval: interval,
	}
}

// Run runs the PostChecker.
func (pc *PostChecker) Run(ctx context.Context, done chan<- error) {
	ticker := time.NewTicker(pc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			done <- ctx.Err()
			return
		case <-ticker.C:
			var p Posts
			// todo: add backoff/retry logic?
			if err := fetchLatestPosts(ctx, pc.client, pc.baseURL, &p); err != nil {
				done <- fmt.Errorf("fetching latest posts: %w", err)
				return
			}
			// todo: filter posts we already checked. We may need to store the last post ID checked somewhere.
			// todo: do something with the posts.
			for _, p := range p.Latest {
				if pc.keywords == "" || p.ContainsKeywords(pc.keywords) {
					fmt.Println("*", p.ID, p.CreatedAt, p.UpdatedAt, p.Username, p.Title)
				}
			}
		}
	}
}

func fetchLatestPosts(ctx context.Context, client http.Client, baseURL string, p *Posts) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("url parse: %w", err)
	}
	u.Path = path.Join(u.Path, "posts.json")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer func() {
		_ = rsp.Body.Close()
	}()

	if err := json.NewDecoder(rsp.Body).Decode(p); err != nil {
		return fmt.Errorf("json decoder: %w", err)
	}
	return nil
}
