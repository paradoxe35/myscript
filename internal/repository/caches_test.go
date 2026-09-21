// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import "testing"

func TestPageLanguageRoundTrip(t *testing.T) {
	db := newDB(t, &Cache{})
	repo := NewCacheRepository(db)

	if got := repo.PageLanguage("p"); got != "" {
		t.Errorf("an unset language reads as %q", got)
	}

	repo.SetPageLanguage("p", "fr")
	repo.SetPageLanguage("p", "de")

	if got := repo.PageLanguage("p"); got != "de" {
		t.Errorf("got %q", got)
	}

	var rows int64
	db.Model(&Cache{}).Count(&rows)
	if rows != 1 {
		t.Errorf("got %d rows for one page, the key should be reused", rows)
	}
}
