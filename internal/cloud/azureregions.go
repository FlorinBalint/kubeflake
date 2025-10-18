package cloud

import (
	"sort"
)

// azureRegions maps Azure region name -> increasing integer (stable order).
// Indices are assigned deterministically. Regions listed in topAzureRegions
// are guaranteed to take the first indices, in sorted(topAzureRegions) order.
var (
	azureRegions = map[string]int{}
)

// topAzureRegions lists the top regions for global coverage.
// They will take the first IDs to ensure a global presence
// even when only 3 bits are used to encode the cluster IDs.
var topAzureRegions = []string{
	"southcentralus",   // South Central US
	"chilecentral",     // Chile Central
	"southeastasia",    // Southeast Asia
	"polandcentral",    // Poland Central
	"japanwest",        // Japan West
	"canadaeast",       // Canada East
	"southafricanorth", // South Africa North
	"uaenorth",         // UAE North
}

// baseAzureRegions contains the baked-in Azure regions.
var baseAzureRegions = []string{
	// Asia Pacific
	"eastasia",      // East Asia
	"southeastasia", // Southeast Asia

	// Australia
	"australiacentral",   // Australia Central
	"australiacentral2",  // Australia Central 2
	"australiaeast",      // Australia East
	"australiasoutheast", // Australia Southeast

	// Austria
	"austriaeast", // Austria East

	// Brazil
	"brazilsouth",     // Brazil South
	"brazilsoutheast", // Brazil Southeast

	// Canada
	"canadacentral", // Canada Central
	"canadaeast",    // Canada East

	// Canary (US)
	"centraluseuap", // Central US EUAP
	"eastus2euap",   // East US 2 EUAP

	// Chile
	"chilecentral", // Chile Central

	// Europe
	"northeurope", // North Europe
	"westeurope",  // West Europe

	// France
	"francecentral", // France Central
	"francesouth",   // France South

	// Germany
	"germanynorth",       // Germany North
	"germanywestcentral", // Germany West Central

	// India
	"centralindia",    // Central India
	"jioindiacentral", // Jio India Central
	"jioindiawest",    // Jio India West
	"southindia",      // South India
	"westindia",       // West India

	// Indonesia
	"indonesiacentral", // Indonesia Central

	// Israel
	"israelcentral", // Israel Central

	// Italy
	"italynorth", // Italy North

	// Japan
	"japaneast", // Japan East
	"japanwest", // Japan West

	// Korea
	"koreacentral", // Korea Central
	"koreasouth",   // Korea South

	// Malaysia
	"malaysiawest", // Malaysia West

	// Mexico
	"mexicocentral", // Mexico Central

	// New Zealand
	"newzealandnorth", // New Zealand North

	// Norway
	"norwayeast", // Norway East
	"norwaywest", // Norway West

	// Poland
	"polandcentral", // Poland Central

	// Qatar
	"qatarcentral", // Qatar Central

	// South Africa
	"southafricanorth", // South Africa North
	"southafricawest",  // South Africa West

	// Spain
	"spaincentral", // Spain Central

	// Stage (US)
	"eastusstg",         // East US STG
	"southcentralusstg", // South Central US STG

	// Sweden
	"swedencentral", // Sweden Central

	// Switzerland
	"switzerlandnorth", // Switzerland North
	"switzerlandwest",  // Switzerland West

	// UAE
	"uaecentral", // UAE Central
	"uaenorth",   // UAE North

	// United Kingdom
	"uksouth", // UK South
	"ukwest",  // UK West

	// United States
	"centralus",      // Central US
	"eastus",         // East US
	"eastus2",        // East US 2
	"northcentralus", // North Central US
	"southcentralus", // South Central US
	"westcentralus",  // West Central US
	"westus",         // West US
	"westus2",        // West US 2
	"westus3",        // West US 3

}

// init builds the index maps using the current data.
func init() {
	rebuildAzureIndices()
}

// AzureRegionIndex returns the index for a region and whether it exists.
func AzureRegionIndex(region string) (int, bool) {
	i, ok := azureRegions[region]
	return i, ok
}

// rebuildAzureIndices rebuilds azureRegions ensuring topAzureRegions come first.
func rebuildAzureIndices() {
	azureRegions = map[string]int{}

	// Create a set of all base regions for validation
	baseSet := make(map[string]struct{}, len(baseAzureRegions))
	for _, r := range baseAzureRegions {
		baseSet[r] = struct{}{}
	}

	// Top regions (that exist in the dataset), sorted
	topRegions := make([]string, 0, len(topAzureRegions))
	for _, r := range topAzureRegions {
		if _, ok := baseSet[r]; ok {
			topRegions = append(topRegions, r)
		}
	}
	sort.Strings(topRegions)

	// Regions: top first, then the rest
	topSet := make(map[string]struct{}, len(topRegions))
	for _, r := range topRegions {
		topSet[r] = struct{}{}
	}

	restRegions := make([]string, 0, len(baseAzureRegions))
	for _, r := range baseAzureRegions {
		if _, ok := topSet[r]; !ok {
			restRegions = append(restRegions, r)
		}
	}
	sort.Strings(restRegions)

	rIdx := 0
	for _, r := range topRegions {
		azureRegions[r] = rIdx
		rIdx++
	}
	for _, r := range restRegions {
		azureRegions[r] = rIdx
		rIdx++
	}
}
