package post_test

import (
	"testing"
	"time"

	"github.com/halkyon/discourse-scanner/internal/post"
)

func TestContainsKeywords(t *testing.T) {
	t.Parallel()

	p := post.Post{
		ID:          123,
		Title:       "has anyone seen my mobile?",
		ContentRaw:  "my mobile PHONE disappeared\ndoes anyone know where it is?",
		ContentHTML: "my mobile PHONE disappeared\ndoes anyone know where it is?",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Username:    "joe",
	}
	tests := []struct {
		name     string
		p        post.Post
		keywords string
		want     bool
	}{
		{
			name:     "empty keywords",
			keywords: "",
			want:     false,
		},
		{
			name:     "post contains single keyword in title",
			keywords: "mobile",
			want:     true,
		},
		{
			name:     "post doesn't contain keyword",
			keywords: "something",
			want:     false,
		},
		{
			name:     "post contains single keyword in content",
			keywords: "disappeared",
			want:     true,
		},
		{
			name:     "post contains multiple keywords in title",
			keywords: "has,seen",
			want:     true,
		},
		{
			name:     "post contains multiple keywords in content",
			keywords: "anyone,know",
			want:     true,
		},
		{
			name:     "post contains keyword in different case",
			keywords: "phone",
			want:     true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := p.ContainsKeywords(tt.keywords); got != tt.want {
				t.Errorf("Post.ContainsKeywords() = %v, want %v", got, tt.want)
			}
		})
	}
}
