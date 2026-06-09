package search

import (
	"context"
	"fmt"

	"github.com/meilisearch/meilisearch-go"
)

const indexName = "wishlist_items"

type Client struct {
	client meilisearch.ServiceManager
}

// Disabled is a search client used when Meilisearch is unavailable at startup.
var Disabled = &Client{}

func New(url, apiKey string) (*Client, error) {
	client := meilisearch.New(url, meilisearch.WithAPIKey(apiKey))
	c := &Client{client: client}
	if err := c.ensureIndex(context.Background()); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) available() bool {
	return c != nil && c.client != nil
}

func (c *Client) ensureIndex(ctx context.Context) error {
	if !c.available() {
		return fmt.Errorf("meilisearch unavailable")
	}
	_, err := c.client.GetIndex(indexName)
	if err == nil {
		return nil
	}
	task, err := c.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        indexName,
		PrimaryKey: "id",
	})
	if err != nil {
		return fmt.Errorf("create meilisearch index: %w", err)
	}
	if _, err := c.client.WaitForTaskWithContext(ctx, task.TaskUID, 0); err != nil {
		return fmt.Errorf("wait for index creation: %w", err)
	}
	filterTask, err := c.client.Index(indexName).UpdateFilterableAttributesWithContext(ctx, &[]string{"user_id", "wishlist_id"})
	if err != nil {
		return err
	}
	_, err = c.client.WaitForTaskWithContext(ctx, filterTask.TaskUID, 0)
	return err
}

type IndexDocument struct {
	ID            string `json:"id"`
	WishlistID    string `json:"wishlist_id"`
	WishlistSlug  string `json:"wishlist_slug"`
	WishlistTitle string `json:"wishlist_title"`
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	UserName      string `json:"user_name"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Brand         string `json:"brand"`
	Notes         string `json:"notes"`
}

func (c *Client) IndexItem(ctx context.Context, doc IndexDocument) error {
	if !c.available() {
		return nil
	}
	task, err := c.client.Index(indexName).AddDocumentsWithContext(ctx, []IndexDocument{doc})
	if err != nil {
		return err
	}
	_, err = c.client.WaitForTaskWithContext(ctx, task.TaskUID, 0)
	return err
}

func (c *Client) DeleteItem(ctx context.Context, id string) error {
	if !c.available() {
		return nil
	}
	task, err := c.client.Index(indexName).DeleteDocumentWithContext(ctx, id)
	if err != nil {
		return err
	}
	_, err = c.client.WaitForTaskWithContext(ctx, task.TaskUID, 0)
	return err
}

func (c *Client) DeleteWishlistItems(ctx context.Context, wishlistID string) error {
	if !c.available() {
		return nil
	}
	filter := fmt.Sprintf("wishlist_id = %q", wishlistID)
	task, err := c.client.Index(indexName).DeleteDocumentsByFilterWithContext(ctx, filter)
	if err != nil {
		return err
	}
	_, err = c.client.WaitForTaskWithContext(ctx, task.TaskUID, 0)
	return err
}

func (c *Client) SearchWishlistItems(ctx context.Context, query string, limit int) ([]WishlistItemHit, error) {
	if !c.available() {
		return []WishlistItemHit{}, nil
	}
	if limit <= 0 {
		limit = 10
	}
	resp, err := c.client.Index(indexName).SearchWithContext(ctx, query, &meilisearch.SearchRequest{Limit: int64(limit)})
	if err != nil {
		return nil, err
	}
	hits := make([]WishlistItemHit, 0, len(resp.Hits))
	for _, hit := range resp.Hits {
		doc, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}
		hits = append(hits, WishlistItemHit{
			ID:            fmt.Sprint(doc["id"]),
			Name:          fmt.Sprint(doc["name"]),
			Type:          fmt.Sprint(doc["type"]),
			Brand:         fmt.Sprint(doc["brand"]),
			Notes:         fmt.Sprint(doc["notes"]),
			Score:         rankingScore(doc),
			UserID:        fmt.Sprint(doc["user_id"]),
			UserName:      fmt.Sprint(doc["user_name"]),
			Username:      fmt.Sprint(doc["username"]),
			WishlistSlug:  fmt.Sprint(doc["wishlist_slug"]),
			WishlistTitle: fmt.Sprint(doc["wishlist_title"]),
		})
	}
	return hits, nil
}

func rankingScore(doc map[string]interface{}) float64 {
	if score, ok := doc["_rankingScore"].(float64); ok {
		return score
	}
	return 0
}
