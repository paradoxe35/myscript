package notion

import (
	"context"
	"log"

	"github.com/jomei/notionapi"
)

type NotionClient struct {
	ctx    context.Context
	Client *notionapi.Client
}

func NewClient(token string) *NotionClient {
	client := notionapi.NewClient(notionapi.Token(token))

	return &NotionClient{
		ctx:    context.Background(),
		Client: client,
	}
}

func (nc *NotionClient) GetPages() []notionapi.Object {
	result, err := nc.Client.Search.Do(nc.ctx, &notionapi.SearchRequest{
		PageSize: 200,
		Filter:   notionapi.SearchFilter{Property: "object", Value: "page"},
	})
	if err != nil {
		log.Printf("Error getting Notion pages: %s", err)
		return []notionapi.Object{}
	}

	return result.Results
}

func (nc *NotionClient) GetPageBlocks(pageID string) []notionapi.Block {
	result, err := nc.Client.Block.GetChildren(
		nc.ctx,
		notionapi.BlockID(pageID),
		&notionapi.Pagination{PageSize: 1000},
	)

	if err != nil {
		log.Printf("Error getting Notion page blocks: %s", err)
		return []notionapi.Block{}
	}

	return result.Results
}
