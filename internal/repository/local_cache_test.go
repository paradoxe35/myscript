// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import "testing"

func TestLocalCacheRoundTrip(t *testing.T) {
	repo := NewLocalCacheRepository(newDB(t, &LocalCache{}))

	type block struct{ Kind string }
	if err := repo.Set("notion_page:1", []block{{Kind: "heading"}}); err != nil {
		t.Fatal(err)
	}

	var blocks []block
	if !repo.Get("notion_page:1", &blocks) {
		t.Fatal("the entry should be found")
	}
	if len(blocks) != 1 || blocks[0].Kind != "heading" {
		t.Errorf("got %+v", blocks)
	}

	if repo.Get("notion_page:2", &blocks) {
		t.Error("a missing key must not read as found")
	}
}

func TestLocalCacheSetReplacesTheValue(t *testing.T) {
	repo := NewLocalCacheRepository(newDB(t, &LocalCache{}))
	repo.Set("k", 1)
	repo.Set("k", 2)

	var value int
	repo.Get("k", &value)
	if value != 2 {
		t.Errorf("got %d", value)
	}
}
