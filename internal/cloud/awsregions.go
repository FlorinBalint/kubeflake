package cloud

import (
	"sort"
)

// AWSRegions maps AWS region name -> increasing integer (stable order).
// Indices are assigned deterministically. Regions listed in topAWSRegions
// are guaranteed to take the first indices, in sorted(topAWSRegions) order.
var AWSRegions = map[string]int{}

// topAWSRegions lists the top regions for global coverage.
// They will take the first IDs to ensure a global presence
// even when only 3 bits are used to encode the cluster IDs.
var topAWSRegions = []string{
	"ap-southeast-2",
	"eu-west-2",
	"us-west-1",
	"ap-east-1",
	"af-south-1",
	"sa-east-1",
	"me-central-1",
	"ca-central-1",
}

// allAWSRegions contains all available AWS regions.
var allAWSRegions = []string{
	"af-south-1",
	"ap-east-1",
	"ap-east-2",
	"ap-northeast-1",
	"ap-northeast-2",
	"ap-northeast-3",
	"ap-south-1",
	"ap-south-2",
	"ap-southeast-1",
	"ap-southeast-2",
	"ap-southeast-3",
	"ap-southeast-4",
	"ap-southeast-5",
	"ap-southeast-6",
	"ap-southeast-7",
	"ca-central-1",
	"ca-west-1",
	"eu-central-1",
	"eu-central-2",
	"eu-north-1",
	"eu-south-1",
	"eu-south-2",
	"eu-west-1",
	"eu-west-2",
	"eu-west-3",
	"il-central-1",
	"me-central-1",
	"me-south-1",
	"mx-central-1",
	"sa-east-1",
	"us-east-1",
	"us-east-2",
	"us-west-1",
	"us-west-2",
}

// init builds the index maps using the current data.
func init() {
	rebuildAWSIndices()
}

// AWSRegionIndex returns the index for a region and whether it exists.
func AWSRegionIndex(region string) (int, bool) {
	i, ok := AWSRegions[region]
	return i, ok
}

// rebuildAWSIndices rebuilds AWSRegions ensuring topAWSRegions come first.
func rebuildAWSIndices() {
	AWSRegions = map[string]int{}

	// Collect regions
	allRegions := make([]string, len(allAWSRegions))
	copy(allRegions, allAWSRegions)
	sort.Strings(allRegions)

	// Top regions (that exist in the dataset), sorted
	topRegions := make([]string, 0, len(topAWSRegions))
	for _, r := range topAWSRegions {
		if hasRegion(allAWSRegions, r) {
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
		AWSRegions[r] = rIdx
		rIdx++
	}
	for _, r := range restRegions {
		AWSRegions[r] = rIdx
		rIdx++
	}
}

func hasRegion(regions []string, want string) bool {
	for _, r := range regions {
		if r == want {
			return true
		}
	}
	return false
}
