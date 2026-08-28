package services_test

import (
	"testing"

	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"github.com/ali/hesab-keepnet/backend/internal/services"
)

func TestNotesDataAttachedLifecycle(t *testing.T) {
	env := openTestEnv(t)
	acct := env.createAccount(t, "ملت", enums.CurrencyRIAL, 0)
	svc := env.Services.Notes

	note, err := svc.Create(env.Ctx, services.CreateNoteInput{
		EntityType: models.NoteEntityBankAccount,
		EntityID:   &acct.ID,
		Body:       "این حساب معمولاً آخر ماه تسویه می‌شود.",
		Tags:       []string{"Log", " مذاکره "},
	})
	if err != nil {
		t.Fatal(err)
	}

	items, total, err := svc.List(env.Ctx, services.NoteFilter{
		EntityType: models.NoteEntityBankAccount,
		EntityID:   &acct.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].Body != note.Body {
		t.Fatalf("list = %v (%d), want single attached note", items, total)
	}
	if items[0].Tags != "log,مذاکره" {
		t.Errorf("tags normalized = %q, want %q", items[0].Tags, "log,مذاکره")
	}

	pinned := true
	if _, err := svc.Update(env.Ctx, note.ID, services.UpdateNoteInput{Pinned: &pinned}); err != nil {
		t.Fatal(err)
	}

	journal, err := svc.Create(env.Ctx, services.CreateNoteInput{
		EntityType: models.NoteEntityJournal,
		Body:       "جمع‌بندی امروز",
		Tags:       []string{"روزانه"},
		Pinned:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if journal.EntityID != nil {
		t.Error("journal notes must not carry an entity id")
	}

	found, _, err := svc.List(env.Ctx, services.NoteFilter{JournalOnly: true, Query: "جمع‌بندی"})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || !found[0].Pinned {
		t.Errorf("journal search = %+v, want pinned entry", found)
	}

	// Tag filter must match the comma-separated tags column.
	tagged, _, err := svc.List(env.Ctx, services.NoteFilter{JournalOnly: true, Tag: "روزانه"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tagged) != 1 {
		t.Errorf("tag filter matched %d notes, want 1", len(tagged))
	}

	if err := svc.Delete(env.Ctx, note.ID); err != nil {
		t.Fatal(err)
	}
	remaining, total2, err := svc.List(env.Ctx, services.NoteFilter{
		EntityType: models.NoteEntityBankAccount,
		EntityID:   &acct.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if total2 != 0 || len(remaining) != 0 {
		t.Errorf("after delete list = %+v (%d), want empty", remaining, total2)
	}
}

func TestNotesValidation(t *testing.T) {
	env := openTestEnv(t)
	svc := env.Services.Notes

	if _, err := svc.Create(env.Ctx, services.CreateNoteInput{
		EntityType: models.NoteEntityJournal,
		Body:       "   ",
	}); err == nil {
		t.Error("empty body must be rejected")
	}
	if _, err := svc.Create(env.Ctx, services.CreateNoteInput{
		EntityType: "HACKER",
		Body:       "x",
	}); err == nil {
		t.Error("unknown entity type must be rejected")
	}
	if _, err := svc.Create(env.Ctx, services.CreateNoteInput{
		EntityType: models.NoteEntityRepresentative,
		Body:       "بدون شناسه",
	}); err == nil {
		t.Error("attached note without entity id must be rejected")
	}
}
