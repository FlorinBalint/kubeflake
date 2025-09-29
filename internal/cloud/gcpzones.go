package cloud

import (
	"sort"
)

// Regions maps GCP region name -> increasing integer (stable order).
// Zones maps GCP zone name -> increasing integer (stable order).
// Indices are assigned deterministically. Zones listed in topRegionZones
// are guaranteed to take the first indices, in sorted(topRegionZones) order.
var (
	gcpRegions = map[string]int{}
	gcpZones   = map[string]int{}
)

// topGcpRegionZones lists the top zones for each region.
// They will take the first IDs to ensure a global presence
// even when only 3 bits are used to encode the cluster IDs.
var topGcpRegionZones = map[string][]string{
	"africa-south1":        {"a"},
	"asia-northeast1":      {"a"},
	"asia-south2":          {"a"},
	"australia-southeast2": {"a"},
	"europe-north1":        {"a"},
	"me-west1":             {"a"},
	"southamerica-east1":   {"a"},
	"us-central1":          {"a"},
}

// baseRegionZones contains the baked-in regions -> zone letters.
var baseGcpRegionZones = map[string][]string{
	// Africa
	"africa-south1": {"a", "b", "c"},

	// Asia
	"asia-east1":      {"a", "b", "c"},
	"asia-east2":      {"a", "b", "c"},
	"asia-northeast1": {"a", "b", "c"},
	"asia-northeast2": {"a", "b", "c"},
	"asia-northeast3": {"a", "b", "c"},
	"asia-south1":     {"a", "b", "c"},
	"asia-south2":     {"a", "b", "c"},
	"asia-southeast1": {"a", "b", "c"},
	"asia-southeast2": {"a", "b", "c"},

	// Australia
	"australia-southeast1": {"a", "b", "c"},
	"australia-southeast2": {"a", "b", "c"},

	// Europe
	"europe-central2":   {"a", "b", "c"},
	"europe-north1":     {"a", "b", "c"},
	"europe-north2":     {"a", "b", "c"},
	"europe-southwest1": {"a", "b", "c"},
	"europe-west1":      {"b", "c", "d"},
	"europe-west10":     {"a", "b", "c"},
	"europe-west12":     {"a", "b", "c"},
	"europe-west2":      {"a", "b", "c"},
	"europe-west3":      {"a", "b", "c"},
	"europe-west4":      {"a", "b", "c"},
	"europe-west6":      {"a", "b", "c"},
	"europe-west8":      {"a", "b", "c"},
	"europe-west9":      {"a", "b", "c"},

	// Middle East
	"me-central1": {"a", "b", "c"},
	"me-central2": {"a", "b", "c"},
	"me-west1":    {"a", "b", "c"},

	// North America
	"northamerica-northeast1": {"a", "b", "c"},
	"northamerica-northeast2": {"a", "b", "c"},
	"northamerica-south1":     {"a", "b", "c"},
	"us-central1":             {"a", "b", "c", "f"},
	"us-east1":                {"b", "c", "d"},
	"us-east4":                {"a", "b", "c"},
	"us-east5":                {"a", "b", "c"},
	"us-south1":               {"a", "b", "c"},
	"us-west1":                {"a", "b", "c"},
	"us-west2":                {"a", "b", "c"},
	"us-west3":                {"a", "b", "c"},
	"us-west4":                {"a", "b", "c"},

	// South America
	"southamerica-east1": {"a", "b", "c"},
	"southamerica-west1": {"a", "b", "c"},
}

// init builds the index maps using the current data.
func init() {
	rebuildIndices()
}

// GCPRegionIndex returns the index for a region and whether it exists.
func GCPRegionIndex(region string) (int, bool) {
	i, ok := gcpRegions[region]
	return i, ok
}

// GCPZoneIndex returns the index for a zone and whether it exists.
func GCPZoneIndex(zone string) (int, bool) {
	i, ok := gcpZones[zone]
	return i, ok
}

// rebuildIndices rebuilds Regions and Zones ensuring topRegionZones come first.
func rebuildIndices() {
	gcpRegions = map[string]int{}
	gcpZones = map[string]int{}

	// Collect regions
	allRegions := make([]string, 0, len(baseGcpRegionZones))
	for r := range baseGcpRegionZones {
		allRegions = append(allRegions, r)
	}
	sort.Strings(allRegions)

	// Top regions (that exist in the dataset), sorted
	topRegions := make([]string, 0, len(topGcpRegionZones))
	for r := range topGcpRegionZones {
		if _, ok := baseGcpRegionZones[r]; ok {
			topRegions = append(topRegions, r)
		}
	}
	sort.Strings(topRegions)

	// Regions: top first, then the rest
	topSet := make(map[string]struct{}, len(topRegions))
	for _, r := range topRegions {
		topSet[r] = struct{}{}
	}
	restRegions := make([]string, 0, len(allRegions))
	for _, r := range allRegions {
		if _, ok := topSet[r]; !ok {
			restRegions = append(restRegions, r)
		}
	}

	rIdx := 0
	for _, r := range topRegions {
		gcpRegions[r] = rIdx
		rIdx++
	}
	for _, r := range restRegions {
		gcpRegions[r] = rIdx
		rIdx++
	}

	// Zones: topGcpRegionZones first (only if present), then remaining zones by region asc, letter asc.
	zIdx := 0
	added := make(map[string]struct{}, 128)

	for _, r := range topRegions {
		letters := append([]string(nil), topGcpRegionZones[r]...)
		sort.Strings(letters)
		for _, l := range letters {
			// Add only if this zone exists in baseGcpRegionZones
			if !hasLetter(baseGcpRegionZones[r], l) {
				continue
			}
			zone := r + "-" + l
			if _, ok := added[zone]; ok {
				continue
			}
			gcpZones[zone] = zIdx
			added[zone] = struct{}{}
			zIdx++
		}
	}

	for _, r := range allRegions {
		letters := append([]string(nil), baseGcpRegionZones[r]...)
		sort.Strings(letters)
		for _, l := range letters {
			zone := r + "-" + l
			if _, ok := added[zone]; ok {
				continue
			}
			gcpZones[zone] = zIdx
			added[zone] = struct{}{}
			zIdx++
		}
	}
}

func hasLetter(letters []string, want string) bool {
	for _, l := range letters {
		if l == want {
			return true
		}
	}
	return false
}
