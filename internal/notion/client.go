package notion

import (
	"context"

	"github.com/jomei/notionapi"
)

type NotionClient struct {
	ctx    context.Context
	Client *notionapi.Client
}

func NewClient(token string) *NotionClient {
	cxt := context.Background()

	client := notionapi.NewClient(notionapi.Token(token))

	return &NotionClient{
		ctx:    cxt,
		Client: client,
	}
}

func (nc *NotionClient) GetPages() []notionapi.Object {
	result, err := nc.Client.Search.Do(nc.ctx, &notionapi.SearchRequest{PageSize: 200})
	if err != nil {
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
		return []notionapi.Block{}
	}

	return result.Results
}
