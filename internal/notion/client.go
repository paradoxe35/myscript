// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package notion

import (
	"context"
	"log/slog"
	"sync"

	"github.com/jomei/notionapi"
)

type NotionClient struct {
	ctx    context.Context
	Client *notionapi.Client
}

type NotionBlock struct {
	Block    notionapi.Block
	Children []*NotionBlock
}

func NewClient(token string) *NotionClient {
	client := notionapi.NewClient(notionapi.Token(token))

	return &NotionClient{
		ctx:    context.Background(),
		Client: client,
	}
}

func (nc *NotionClient) GetPages() ([]notionapi.Object, error) {
	result, err := nc.Client.Search.Do(nc.ctx, &notionapi.SearchRequest{
		PageSize: 200,
		Filter:   notionapi.SearchFilter{Property: "object", Value: "page"},
	})

	if err != nil {
		slog.Error("Error getting Notion pages", "error", err)
		return []notionapi.Object{}, err
	}

	return result.Results, nil
}

func (nc *NotionClient) fetchBlocks(pageID string) ([]*NotionBlock, error) {
	blocks := []*NotionBlock{}
	result, err := nc.Client.Block.GetChildren(
		nc.ctx,
		notionapi.BlockID(pageID),
		&notionapi.Pagination{PageSize: 1000},
	)

	if err != nil {
		slog.Error("Error getting Notion page blocks", "error", err)
		return blocks, err
	}

	var wg sync.WaitGroup

	for _, block := range result.Results {
		newBlock := &NotionBlock{Block: block, Children: []*NotionBlock{}}

		if block.GetHasChildren() {
			wg.Add(1)
			go func(blockID string) {
				defer wg.Done()
				newBlock.Children, _ = nc.fetchBlocks(blockID)
			}(block.GetID().String())
		}

		blocks = append(blocks, newBlock)
	}

	wg.Wait()

	return blocks, nil
}

func (nc *NotionClient) GetPageBlocks(pageID string) ([]*NotionBlock, error) {
	return nc.fetchBlocks(pageID)
}
