// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package synchronizer

import (
	"myscript/internal/repository"
	"testing"
	"time"
)

func change(id uint, table, row string, at time.Time) repository.ChangeLog {
	log := repository.ChangeLog{TableName: table, RowID: row, ChangeID: table + ":" + row}
	log.ID = id
	log.UpdatedAt = at
	return log
}

func TestChangesForOneRowStayInOrder(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	groups := groupChangesByRow([]repository.ChangeLog{
		change(3, "pages", "a", base.Add(2*time.Minute)),
		change(1, "pages", "a", base),
		change(2, "pages", "a", base.Add(time.Minute)),
	})

	if len(groups) != 1 {
		t.Fatalf("got %d groups, one row is one group", len(groups))
	}
	for i, want := range []uint{1, 2, 3} {
		if groups[0][i].ID != want {
			t.Errorf("position %d holds change %d, want %d", i, groups[0][i].ID, want)
		}
	}
}

func TestDifferentRowsAreSeparateGroups(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	groups := groupChangesByRow([]repository.ChangeLog{
		change(1, "pages", "a", base),
		change(2, "pages", "b", base),
		change(3, "configs", "a", base),
	})

	if len(groups) != 3 {
		t.Fatalf("got %d groups, want 3: rows are independent", len(groups))
	}
}

// Rows carrying the same timestamp must still have a stable order, or a retry
// could upload them the other way round.
func TestAnEqualTimestampFallsBackToInsertionOrder(t *testing.T) {
	at := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	groups := groupChangesByRow([]repository.ChangeLog{
		change(9, "pages", "a", at),
		change(4, "pages", "a", at),
	})

	if groups[0][0].ID != 4 || groups[0][1].ID != 9 {
		t.Errorf("got %d then %d, want 4 then 9", groups[0][0].ID, groups[0][1].ID)
	}
}

func TestAChangeWithoutAnUpdateTimeUsesItsCreationTime(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	older := repository.ChangeLog{TableName: "pages", RowID: "a"}
	older.ID = 2
	older.CreatedAt = base

	newer := change(1, "pages", "a", base.Add(time.Hour))

	groups := groupChangesByRow([]repository.ChangeLog{newer, older})

	if groups[0][0].ID != 2 {
		t.Errorf("the older change should go first, got %d", groups[0][0].ID)
	}
}

func TestGroupingAnEmptyListIsSafe(t *testing.T) {
	if groups := groupChangesByRow(nil); len(groups) != 0 {
		t.Errorf("got %d groups", len(groups))
	}
}

func TestGroupingDoesNotReorderTheCaller(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	changes := []repository.ChangeLog{
		change(2, "pages", "a", base.Add(time.Minute)),
		change(1, "pages", "a", base),
	}

	groupChangesByRow(changes)

	if changes[0].ID != 2 {
		t.Error("the caller's slice should not be sorted in place")
	}
}
