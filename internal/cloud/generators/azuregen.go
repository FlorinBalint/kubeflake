package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

var (
	topRegionsFlag = flag.String("top-regions", "", "Override top regions in comma-separated format (exactly 4 or 8 regions, e.g., 'eastus,westeurope,southeastasia,australiaeast'). If not provided, uses default regions.")
)

// AzureLocation represents an Azure location from the az CLI output
type AzureLocation struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	RegionalName string `json:"regionalDisplayName"`
	Metadata     struct {
		RegionType     string `json:"regionType"`
		RegionCategory string `json:"regionCategory"`
		Geography      string `json:"geography"`
	} `json:"metadata"`
}

// RegionInfo represents a region with its display name
type RegionInfo struct {
	Name        string
	DisplayName string
}

// RegionConfig represents a top region configuration with display name
type RegionConfig struct {
	Name        string
	DisplayName string
}

// Config represents the template configuration
type Config struct {
	AllRegions map[string][]RegionInfo
	TopRegions []RegionConfig
}

// TemplateData represents the data passed to the template
type TemplateData struct {
	Config Config
}

// parseTopRegions parses the top regions override flag format: "region1,region2,..."
func parseTopRegions(topRegionsFlag string) []string {
	if topRegionsFlag == "" {
		return nil
	}

	regions := strings.Split(topRegionsFlag, ",")
	result := make([]string, 0, len(regions))
	for _, region := range regions {
		trimmed := strings.TrimSpace(region)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GenerateAzureRegionsFile runs az command and generates azureregions.go file
func GenerateAzureRegionsFile(customTopRegions []string) error {
	// Run az account list-locations command
	locations, err := getAzureLocations()
	if err != nil {
		return fmt.Errorf("failed to get Azure locations: %w", err)
	}

	// Process locations into regions
	config := processLocationsIntoConfig(locations, customTopRegions)

	// Generate the file from template
	err = generateFileFromTemplate(config)
	if err != nil {
		return fmt.Errorf("failed to generate file from template: %w", err)
	}

	fmt.Println("Successfully generated azureregions.go")
	return nil
}

// getAzureLocations runs az command and parses the JSON output
func getAzureLocations() ([]AzureLocation, error) {
	cmd := exec.Command("az", "account", "list-locations", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run az command: %w", err)
	}

	var locations []AzureLocation
	err = json.Unmarshal(output, &locations)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w", err)
	}

	return locations, nil
}

// processLocationsIntoConfig converts locations into the config structure expected by the template
func processLocationsIntoConfig(locations []AzureLocation, customTopRegions []string) Config {
	regionMap := make(map[string]string) // region name -> display name
	geographyMap := make(map[string][]RegionInfo)

	// Filter and group regions by geography
	for _, location := range locations {
		// Only include physical regions (both Recommended and Other categories)
		if location.Metadata.RegionType != "Physical" {
			continue
		}
		// Skip logical regions like staging environments
		if location.Metadata.RegionCategory != "Recommended" && location.Metadata.RegionCategory != "Other" {
			continue
		}

		regionMap[location.Name] = location.DisplayName

		regionInfo := RegionInfo{
			Name:        location.Name,
			DisplayName: location.DisplayName,
		}

		geography := classifyRegionByGeography(location)
		geographyMap[geography] = append(geographyMap[geography], regionInfo)
	}

	// Sort regions within each geography
	for geography := range geographyMap {
		sort.Slice(geographyMap[geography], func(i, j int) bool {
			return geographyMap[geography][i].Name < geographyMap[geography][j].Name
		})
	}

	// Select top regions
	topRegions := selectTopRegions(regionMap, customTopRegions)

	return Config{
		AllRegions: geographyMap,
		TopRegions: topRegions,
	}
}

// classifyRegionByGeography classifies regions by their geography
func classifyRegionByGeography(location AzureLocation) string {
	geography := location.Metadata.Geography
	if geography == "" {
		return "Other"
	}
	return geography
}

// selectTopRegions validates and applies the custom top regions
func selectTopRegions(regionMap map[string]string, customTopRegions []string) []RegionConfig {
	topRegions := make([]RegionConfig, 0, len(customTopRegions))

	// Apply and validate custom top regions
	for _, region := range customTopRegions {
		if displayName, exists := regionMap[region]; exists {
			topRegions = append(topRegions, RegionConfig{
				Name:        region,
				DisplayName: displayName,
			})
		} else {
			fmt.Printf("Warning: Region '%s' not found in available regions\n", region)
		}
	}

	return topRegions
}

// generateFileFromTemplate generates the azurezones.go file using the template
func generateFileFromTemplate(config Config) error {
	// Read the template file (relative to current directory)
	templatePath := "templates/azureregions.go.template"
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse the template
	tmpl, err := template.New("azureregions").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output directory if it doesn't exist (relative to project root)
	outputDir := "../"
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	outputPath := filepath.Join(outputDir, "azureregions.go")
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Execute template
	data := TemplateData{Config: config}
	err = tmpl.Execute(outputFile, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func main() {
	flag.Parse()
	var customTopRegions []string

	if *topRegionsFlag == "" {
		// Use default hardcoded regions (8 regions for global coverage)
		customTopRegions = []string{
			"southcentralus",
			"chilecentral",
			"southeastasia",
			"polandcentral",
			"japanwest",
			"canadaeast",
			"southafricanorth",
			"uaenorth",
		}
		fmt.Printf("Using default 8 top regions: %v\n", customTopRegions)
	} else {
		// Parse the custom top regions
		customTopRegions = parseTopRegions(*topRegionsFlag)

		// Validate that exactly 2, 4 or 8 top regions are provided
		if len(customTopRegions) != 2 && len(customTopRegions) != 4 && len(customTopRegions) != 8 {
			log.Fatalf("Error: You must provide exactly 2, 4 or 8 top regions, but you provided %d regions.\nProvided regions: %v", len(customTopRegions), customTopRegions)
		}

		fmt.Printf("Using %d custom top regions: %v\n", len(customTopRegions), customTopRegions)
	}

	err := GenerateAzureRegionsFile(customTopRegions)
	if err != nil {
		log.Fatalf("Error generating Azure regions file: %v", err)
	}
}
