package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/zeebo/errs"
)

const (
	discourseURLFlagName   = "discourse-url"
	timeoutFlagName        = "timeout"
	filterKeywordsFlagName = "filter-keywords"
)

var (
	discourseURL   = flag.String(discourseURLFlagName, "", "URL to Discourse instance")
	timeout        = flag.Duration(timeoutFlagName, time.Second*10, "Timeout after fetching post data")
	filterKeywords = flag.String(filterKeywordsFlagName, "", "Comma separated list of keywords to filter posts by")
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
		log.Fatal(err)
	}
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

	var p Posts
	if err := fetchPosts(ctx, client, *discourseURL, &p); err != nil {
		return err
	}

	// todo: filter by keywords if provided.
	// todo: decide if we want to check all posts including replies, or just
	// the first post in a thread.
	// todo: run in a loop, and only check posts that haven't already been checked

	for _, p := range p.LatestPosts {
		fmt.Println(p.ID, p.CreatedAt, p.UpdatedAt, p.Username, p.Title)
	}

	return nil
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
	defer func() {
		_ = rsp.Body.Close()
	}()

	return json.NewDecoder(rsp.Body).Decode(p)
}

func joinURL(baseURL, endpoint string) (string, error) {
	url, err := url.Parse(*discourseURL)
	if err != nil {
		return "", err
	}
	url.Path = path.Join(url.Path, endpoint)
	return url.String(), nil
}
