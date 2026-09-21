// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import "testing"

func TestDeviceSettingsDefaultToLocalTranscription(t *testing.T) {
	repo := NewDeviceSettingsRepository(newDB(t, &DeviceSettings{}))

	if got := repo.Get().TranscriberSource; got != TranscriberLocal {
		t.Errorf("TranscriberSource = %q", got)
	}

	repo.Save(DeviceSettings{MicInputDevice: "Built-in"})
	if got := repo.Get().TranscriberSource; got != TranscriberLocal {
		t.Errorf("a save without a source should still read as local, got %q", got)
	}
}

func TestDeviceSettingsAreASingleRow(t *testing.T) {
	db := newDB(t, &DeviceSettings{})
	repo := NewDeviceSettingsRepository(db)

	model := "whisper-small"
	repo.Save(DeviceSettings{TranscriberSource: TranscriberRemote, SpeechModelID: &model})
	saved := repo.Save(DeviceSettings{TranscriberSource: TranscriberWitAI, AIProvider: "gemini"})

	var rows int64
	db.Model(&DeviceSettings{}).Count(&rows)
	if rows != 1 {
		t.Errorf("got %d rows", rows)
	}
	if saved.TranscriberSource != TranscriberWitAI || saved.AIProvider != "gemini" {
		t.Errorf("got %+v", saved)
	}
	if saved.SpeechModelID != nil {
		t.Error("Save replaces the whole row")
	}
}
