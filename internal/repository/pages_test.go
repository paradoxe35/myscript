// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"encoding/json"
	"testing"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func newPages(t *testing.T) (repo *PageRepository, mainDB, unsynced *gorm.DB) {
	t.Helper()

	mainDB, unsynced = newStores(t)
	return NewPageRepository(mainDB, unsynced), mainDB, unsynced
}

func seedPage(t *testing.T, repo *PageRepository) *Page {
	t.Helper()

	return repo.SavePage(&Page{
		Title:       "Draft",
		HtmlContent: "<p>kept</p>",
		Blocks:      datatypes.JSON(`[{"type":"paragraph"}]`),
	})
}

// The sidebar lists pages without their content, so a rename must not save
// the listed row back.
func TestUpdatingTheTitleLeavesTheContentIntact(t *testing.T) {
	repo, _, _ := newPages(t)
	page := seedPage(t, repo)

	renamed := repo.UpdatePageTitle(page.ID, "Final")

	if renamed.Title != "Final" {
		t.Errorf("title = %q", renamed.Title)
	}
	if renamed.HtmlContent != "<p>kept</p>" || string(renamed.Blocks) != `[{"type":"paragraph"}]` {
		t.Errorf("the content was lost: %+v", renamed)
	}
}

func TestATitleUpdateLogsTheWholePage(t *testing.T) {
	repo, _, _ := newPages(t)
	changes := newDB(t, &ChangeLog{})
	SetUnSyncedDB(changes)
	t.Cleanup(func() { SetUnSyncedDB(nil) })

	page := seedPage(t, repo)
	repo.UpdatePageTitle(page.ID, "Final")

	var change ChangeLog
	changes.First(&change, "row_id = ?", page.ID)

	var logged Page
	if err := json.Unmarshal(change.NewData, &logged); err != nil {
		t.Fatal(err)
	}
	if logged.Title != "Final" || logged.HtmlContent != "<p>kept</p>" {
		t.Errorf("the other machine would receive %+v", logged)
	}
}

func TestExpandedComesFromThisMachine(t *testing.T) {
	repo, mainDB, _ := newPages(t)
	page := seedPage(t, repo)

	if err := repo.SetExpanded(page.ID, true); err != nil {
		t.Fatal(err)
	}

	if !repo.GetPage(page.ID).Expanded {
		t.Error("GetPage should carry the folder state")
	}
	pages := repo.GetPages()
	if len(pages) != 1 || !pages[0].Expanded {
		t.Error("GetPages should carry the folder state")
	}

	if mainDB.Migrator().HasColumn(&Page{}, "Expanded") {
		t.Error("the folder state must not live in the synced table")
	}
}

func TestSavePageDoesNotTouchTheFolderState(t *testing.T) {
	repo, _, _ := newPages(t)
	page := seedPage(t, repo)
	repo.SetExpanded(page.ID, true)

	page.Expanded = false
	saved := repo.SavePage(page)

	if !saved.Expanded {
		t.Error("a full save must not close the folder")
	}
}

func TestDeletingAPageDropsItsState(t *testing.T) {
	repo, _, unsynced := newPages(t)
	page := seedPage(t, repo)
	repo.SetExpanded(page.ID, true)

	repo.DeletePage(page.ID)

	var states int64
	unsynced.Model(&PageState{}).Count(&states)
	if states != 0 {
		t.Errorf("%d state rows survived the page", states)
	}
}
