package repository

import (
	"testing"

	"github.com/amaumene/snowfinder_common/models"
)

func TestScopedResortSlug(t *testing.T) {
	t.Parallel()

	got := scopedResortSlug("mount-foo", " Nagano ", "Shiga Kogen")
	if got != "mount-foo--nagano--shiga kogen" {
		t.Fatalf("scopedResortSlug() = %q", got)
	}
}

func TestResolvePersistedResortRecord_UsesExistingSlugForSameIdentity(t *testing.T) {
	t.Parallel()

	resort := &models.Resort{Slug: "mount-foo", Prefecture: "nagano", Region: "north"}
	existing := &resortIdentityRecord{ID: "resort-1", Slug: "mount-foo", Prefecture: "Nagano", Region: "North"}

	got := resolvePersistedResortRecord(resort, existing, nil)
	if got.ID != "resort-1" || got.Slug != "mount-foo" {
		t.Fatalf("resolvePersistedResortRecord() = %+v", got)
	}
}

func TestResolvePersistedResortRecord_ScopesSlugOnCollision(t *testing.T) {
	t.Parallel()

	resort := &models.Resort{Slug: "mount-foo", Prefecture: "gifu", Region: "west"}
	existing := &resortIdentityRecord{ID: "resort-1", Slug: "mount-foo", Prefecture: "nagano", Region: "north"}

	got := resolvePersistedResortRecord(resort, existing, nil)
	if got.ID != "" {
		t.Fatalf("expected new record, got id %q", got.ID)
	}
	if got.Slug != "mount-foo--gifu--west" {
		t.Fatalf("resolvePersistedResortRecord() slug = %q", got.Slug)
	}
}

func TestResolvePersistedResortRecord_ReusesExistingScopedSlug(t *testing.T) {
	t.Parallel()

	resort := &models.Resort{Slug: "mount-foo", Prefecture: "gifu", Region: "west"}
	existing := &resortIdentityRecord{ID: "resort-1", Slug: "mount-foo", Prefecture: "nagano", Region: "north"}
	existingScoped := &resortIdentityRecord{ID: "resort-2", Slug: "mount-foo--gifu--west", Prefecture: "gifu", Region: "west"}

	got := resolvePersistedResortRecord(resort, existing, existingScoped)
	if got.ID != "resort-2" || got.Slug != "mount-foo--gifu--west" {
		t.Fatalf("resolvePersistedResortRecord() = %+v", got)
	}
}

func TestResolvePersistedResortRecordOrError_ReturnsErrorForUnrelatedScopedCollision(t *testing.T) {
	t.Parallel()

	resort := &models.Resort{Name: "New Resort", Slug: "mount-foo", Prefecture: "gifu", Region: "west"}
	existing := &resortIdentityRecord{ID: "resort-1", Name: "Existing Resort", Slug: "mount-foo", Prefecture: "nagano", Region: "north"}
	existingScoped := &resortIdentityRecord{ID: "resort-2", Name: "Other Resort", Slug: "mount-foo--gifu--west", Prefecture: "gifu", Region: "south"}

	got, err := resolvePersistedResortRecordOrError(resort, existing, existingScoped)
	if err == nil {
		t.Fatal("expected error")
	}
	if got != nil {
		t.Fatalf("expected nil record, got %+v", got)
	}
}

func TestResolvePersistedResortRecordOrError_ReturnsErrorForSlugCollisionAcrossPrefectures(t *testing.T) {
	t.Parallel()

	resort := &models.Resort{Name: "Second Resort", Slug: "mount-foo", Prefecture: "gifu", Region: "west"}
	existing := &resortIdentityRecord{ID: "resort-1", Name: "First Resort", Slug: "mount-foo", Prefecture: "nagano", Region: "north"}
	existingScoped := &resortIdentityRecord{ID: "resort-2", Name: "Scoped Resort", Slug: "mount-foo--gifu--west", Prefecture: "toyama", Region: "west"}

	got, err := resolvePersistedResortRecordOrError(resort, existing, existingScoped)
	if err == nil {
		t.Fatal("expected error")
	}
	if got != nil {
		t.Fatalf("expected nil record, got %+v", got)
	}
}

func TestResolvePersistedResortRecord_EmptySlug(t *testing.T) {
	t.Parallel()

	resort := &models.Resort{Name: "Nameless Slug Resort", Slug: "", Prefecture: "nagano", Region: "north"}

	got := resolvePersistedResortRecord(resort, nil, nil)
	if got == nil {
		t.Fatal("expected record")
	}
	if got.Slug != "" {
		t.Fatalf("resolvePersistedResortRecord() slug = %q, want empty", got.Slug)
	}
}

func TestResolvePersistedResortRecord_EmptyName(t *testing.T) {
	t.Parallel()

	resort := &models.Resort{Name: "", Slug: "mount-foo", Prefecture: "nagano", Region: "north"}
	existing := &resortIdentityRecord{ID: "resort-1", Name: "Existing Resort", Slug: "mount-foo", Prefecture: "Nagano", Region: "North"}

	got := resolvePersistedResortRecord(resort, existing, nil)
	if got == nil {
		t.Fatal("expected record")
	}
	if got.ID != "resort-1" || got.Slug != "mount-foo" {
		t.Fatalf("resolvePersistedResortRecord() = %+v", got)
	}
}
