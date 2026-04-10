package repository

import (
	"fmt"
	"strings"

	"github.com/amaumene/snowfinder_common/models"
)

type resortIdentityRecord struct {
	ID         string
	Slug       string
	Name       string
	Prefecture string
	Region     string
}

func sameResortIdentity(a, b *resortIdentityRecord) bool {
	if a == nil || b == nil {
		return false
	}

	return normalizeIdentityPart(a.Prefecture) == normalizeIdentityPart(b.Prefecture) &&
		normalizeIdentityPart(a.Region) == normalizeIdentityPart(b.Region)
}

func resortIdentityFromModel(resort *models.Resort) *resortIdentityRecord {
	if resort == nil {
		return nil
	}

	return &resortIdentityRecord{
		ID:         resort.ID,
		Slug:       resort.Slug,
		Name:       resort.Name,
		Prefecture: resort.Prefecture,
		Region:     resort.Region,
	}
}

func normalizeIdentityPart(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func scopedResortSlug(slug, prefecture, region string) string {
	parts := []string{strings.TrimSpace(slug)}

	if normalizedPrefecture := normalizeIdentityPart(prefecture); normalizedPrefecture != "" {
		parts = append(parts, normalizedPrefecture)
	}

	if normalizedRegion := normalizeIdentityPart(region); normalizedRegion != "" {
		parts = append(parts, normalizedRegion)
	}

	return strings.Join(parts, "--")
}

func resolvePersistedResortRecord(resort *models.Resort, existingBySlug, existingByScopedSlug *resortIdentityRecord) *resortIdentityRecord {
	current := resortIdentityFromModel(resort)

	if sameResortIdentity(existingBySlug, current) {
		return &resortIdentityRecord{ID: existingBySlug.ID, Slug: existingBySlug.Slug}
	}

	if sameResortIdentity(existingByScopedSlug, current) {
		return &resortIdentityRecord{ID: existingByScopedSlug.ID, Slug: existingByScopedSlug.Slug}
	}

	if existingBySlug != nil && existingByScopedSlug != nil {
		return nil
	}

	if existingBySlug != nil {
		return &resortIdentityRecord{Slug: scopedResortSlug(resort.Slug, resort.Prefecture, resort.Region)}
	}

	return &resortIdentityRecord{Slug: resort.Slug}
}

func resolvePersistedResortRecordOrError(resort *models.Resort, existingBySlug, existingByScopedSlug *resortIdentityRecord) (*resortIdentityRecord, error) {
	record := resolvePersistedResortRecord(resort, existingBySlug, existingByScopedSlug)
	if record != nil {
		return record, nil
	}

	return nil, fmt.Errorf(
		"slug collision: cannot disambiguate %q for resort %q (prefecture %q) from existing resort %q",
		resort.Slug,
		resort.Name,
		resort.Prefecture,
		existingBySlug.Name,
	)
}
