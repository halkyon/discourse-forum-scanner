package postchecker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/halkyon/discourse-forum-scanner/internal/post"
	"github.com/zeebo/errs"
)

const requestTimeoutSeconds = 10

// ErrFetchingPost occurs when a request fails to retrieve post data. This is most likely due to
// an invalid URL, or there's network connectivity issues.
var ErrFetchingPost = errs.New("error fetching post")

// errorFunc is a callback for when PostChecker encounters an error.
type errFunc func(error)

// Posts represents a response containing a number of posts.
type Posts struct {
	Latest []post.Post `json:"latest_posts"`
}

// PostChecker represents a way of checking forum posts for keywords.
type PostChecker struct {
	client   http.Client
	url      string
	keywords string
	interval time.Duration
	errFunc  errFunc
}

// New returns a new instance of PostChecker.
func New(url, keywords string, interval time.Duration, f errFunc) *PostChecker {
	return &PostChecker{
		client:   http.Client{Timeout: requestTimeoutSeconds * time.Second},
		url:      url,
		keywords: keywords,
		interval: interval,
		errFunc:  f,
	}
}

// Run runs the PostChecker.
func (pc *PostChecker) Run(ctx context.Context, done chan<- error) {
	ticker := time.NewTicker(pc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			done <- nil
			return
		case <-ticker.C:
			var p Posts
			// todo: add backoff/retry logic?
			if err := fetchLatestPosts(ctx, pc.client, pc.url, &p); err != nil {
				if pc.errFunc != nil {
					pc.errFunc(fmt.Errorf("%w: %s", ErrFetchingPost, err))
				}
				continue
			}
			// todo: filter posts we already checked. We may need to store the last post ID checked somewhere.
			// todo: do something with the posts.
			for _, p := range p.Latest {
				if p.ContainsKeywords(pc.keywords) {
					fmt.Println("*", p.ID, p.CreatedAt, p.UpdatedAt, p.Username, p.Title)
				}
			}
		}
	}
}

func fetchLatestPosts(ctx context.Context, client http.Client, baseURL string, p *Posts) error {
	url, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	url.Path = path.Join(url.Path, "posts.json")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	return errs.Combine(json.NewDecoder(rsp.Body).Decode(p), rsp.Body.Close())
}
