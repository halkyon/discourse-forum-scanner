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

var (
	discourseURL   = flag.String("discourss-url", "", "URL to Discourse instance")
	timeout        = flag.Duration("timeout", time.Second*10, "Timeout after fetching post data")
	filterKeywords = flag.String("filter-keywords", "", "Comma separated list of keywords to filter posts by")
)

// PostsResponse represents a response containing a number of forum posts.
type PostsResponse struct {
	Posts []Post `json:"latest_posts"`
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
	validate("discourse-url", *discourseURL)
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

	postsURL, err := joinURL(*discourseURL, "posts.json")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, postsURL, nil)
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

	var pr PostsResponse
	if err := json.NewDecoder(rsp.Body).Decode(&pr); err != nil {
		return err
	}

	// todo: filter by keywords if provided.
	// todo: run in a loop, and only check posts that haven't already been checked

	for _, p := range pr.Posts {
		fmt.Println(p.ID, p.CreatedAt, p.UpdatedAt, p.Username, p.Title)
	}

	return nil
}

func joinURL(baseURL, endpoint string) (string, error) {
	url, err := url.Parse(*discourseURL)
	if err != nil {
		return "", err
	}
	url.Path = path.Join(url.Path, endpoint)
	return url.String(), nil
}
