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
