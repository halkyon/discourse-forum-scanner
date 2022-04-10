package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/zeebo/errs"
)

const (
	discourseURLFlagName  = "discourse-url"
	timeoutFlagName       = "timeout"
	checkIntervalFlagName = "check-interval"
	// filterKeywordsFlagName = "filter-keywords".
)

var (
	discourseURL  = flag.String(discourseURLFlagName, "", "URL to Discourse instance")
	timeout       = flag.Duration(timeoutFlagName, 10*time.Second, "Timeout after fetching post data")
	checkInterval = flag.Duration(checkIntervalFlagName, 10*time.Minute, "Interval betweenn fetching posts")
	// filterKeywords = flag.String(filterKeywordsFlagName, "", "Comma separated list of keywords to filter posts by").
)

// Posts represents a response containing a number of posts.
type Posts struct {
	LatestPosts []Post `json:"latest_posts"`
}

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

func validateFlags() (err error) {
	validate := func(name, value string) {
		if value == "" {
			err = errs.Combine(err, fmt.Errorf("flag %s is empty", name))
		}
	}
	validate(discourseURLFlagName, *discourseURL)
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("fatal: %+v\n", err)
		os.Exit(1)
	}
	fmt.Println("finished")
}

func run() error {
	flag.Parse()
	if err := validateFlags(); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	client := http.Client{
		Timeout: *timeout,
	}

	ticker := time.NewTicker(*checkInterval)
	defer ticker.Stop()

	done := make(chan error, 1)

	go func(ctx context.Context, client http.Client, url string, ticker *time.Ticker, done chan<- error) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("signal received, finishing")
				done <- nil
				return
			case <-ticker.C:
				var p Posts
				if err := fetchPosts(ctx, client, url, &p); err != nil {
					fmt.Fprintf(os.Stderr, "error fetching posts: %s\n", err)
					continue
				}
				// todo: decide if we want to check all posts including replies, or just
				// the first post in a thread.

				// todo: filter by keywords if provided.
				// todo: only check posts that haven't been checked. Can we pass some filter query param
				// to posts.json to only fetch posts we haven't seen?
				for _, p := range p.LatestPosts {
					fmt.Println(p.ID, p.CreatedAt, p.UpdatedAt, p.Username, p.Title)
				}
			}
		}
	}(ctx, client, *discourseURL, ticker, done)

	return <-done
}

func fetchPosts(ctx context.Context, client http.Client, baseURL string, p *Posts) error {
	url, err := joinURL(baseURL, "posts.json")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

func joinURL(baseURL, endpoint string) (string, error) {
	url, err := url.Parse(*discourseURL)
	if err != nil {
		return "", err
	}
	url.Path = path.Join(url.Path, endpoint)
	return url.String(), nil
}
