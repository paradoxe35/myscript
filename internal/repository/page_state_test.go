// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import "testing"

func TestReadProgressAndFolderStateShareARowWithoutClobbering(t *testing.T) {
	repo := NewPageStateRepository(newDB(t, &PageState{}))

	if err := repo.SaveReadProgress("p", 12, 300); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetExpanded("p", true); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveReadProgress("p", 40, 300); err != nil {
		t.Fatal(err)
	}

	state := repo.Get("p")
	if !state.Expanded {
		t.Error("saving progress closed the folder")
	}
	if state.ReadWord != 40 || state.ReadTotal != 300 {
		t.Errorf("progress = %d/%d", state.ReadWord, state.ReadTotal)
	}
}

func TestMissingStateReadsAsDefaults(t *testing.T) {
	repo := NewPageStateRepository(newDB(t, &PageState{}))

	state := repo.Get("nowhere")
	if state.Expanded || state.ReadWord != 0 || state.ReadTotal != 0 {
		t.Errorf("got %+v", state)
	}
	if len(repo.All()) != 0 {
		t.Error("All should be empty")
	}
}

func TestAllIsKeyedByPage(t *testing.T) {
	repo := NewPageStateRepository(newDB(t, &PageState{}))
	repo.SetExpanded("a", true)
	repo.SetExpanded("b", false)

	all := repo.All()
	if len(all) != 2 || !all["a"].Expanded || all["b"].Expanded {
		t.Errorf("got %+v", all)
	}
}
