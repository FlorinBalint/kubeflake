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
	topZonesFlag = flag.String("top-zones", "", "Override top zones in format 'region1:zone1,region2:zone2' (exactly 4 or 8 regions, e.g., 'us-central1:b,europe-west1:c,asia-east1:a,australia-southeast1:a'). If not provided, uses default regions.")
)

// GCPZone represents a GCP zone from the gcloud output
type GCPZone struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Status string `json:"status"`
}

// RegionInfo represents a region with its zones
type RegionInfo struct {
	Name  string
	Zones []string
}

// ZoneConfig represents the top zone for a region
type ZoneConfig struct {
	Id string
}

// Config represents the template configuration
type Config struct {
	AllRegions map[string][]RegionInfo
	TopZones   map[string]ZoneConfig
}

// TemplateData represents the data passed to the template
type TemplateData struct {
	Config Config
}

// parseTopZones parses the top zones override flag format: "region1:zone1,region2:zone2"
func parseTopZones(topZonesFlag string) map[string]string {
	topZones := make(map[string]string)
	if topZonesFlag == "" {
		return topZones
	}

	pairs := strings.Split(topZonesFlag, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) == 2 {
			region := strings.TrimSpace(parts[0])
			zone := strings.TrimSpace(parts[1])
			topZones[region] = zone
		}
	}
	return topZones
}

// GenerateGCPZonesFile runs gcloud command and generates gcpzones.go file
func GenerateGCPZonesFile(customTopZones map[string]string) error {
	// Run gcloud compute zones list command
	zones, err := getGCPZones()
	if err != nil {
		return fmt.Errorf("failed to get GCP zones: %w", err)
	}

	// Process zones into regions
	config := processZonesIntoConfig(zones, customTopZones)

	// Generate the file from template
	err = generateFileFromTemplate(config)
	if err != nil {
		return fmt.Errorf("failed to generate file from template: %w", err)
	}

	fmt.Println("Successfully generated gcpzones.go")
	return nil
}

// getGCPZones runs gcloud command and parses the JSON output
func getGCPZones() ([]GCPZone, error) {
	cmd := exec.Command("gcloud", "compute", "zones", "list", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run gcloud command: %w", err)
	}

	var zones []GCPZone
	err = json.Unmarshal(output, &zones)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w", err)
	}

	return zones, nil
}

// processZonesIntoConfig converts zones into the config structure expected by the template
func processZonesIntoConfig(zones []GCPZone, customTopZones map[string]string) Config {
	regionMap := make(map[string][]string)
	continentMap := make(map[string][]RegionInfo)

	// Group zones by region
	for _, zone := range zones {
		if zone.Status != "UP" {
			continue // Skip unavailable zones
		}

		// Extract zone letter (last character after last hyphen)
		parts := strings.Split(zone.Name, "-")
		if len(parts) < 2 {
			continue
		}
		zoneLetter := parts[len(parts)-1]

		// Extract region name from URL if it's a URL, otherwise use as-is
		regionName := zone.Region
		if strings.Contains(regionName, "/") {
			urlParts := strings.Split(regionName, "/")
			regionName = urlParts[len(urlParts)-1] // Get the last part after the last slash
		}

		regionMap[regionName] = append(regionMap[regionName], zoneLetter)
	}

	// Group regions by continent (rough classification based on naming)
	for region, zoneLetters := range regionMap {
		sort.Strings(zoneLetters)

		regionInfo := RegionInfo{
			Name:  region,
			Zones: zoneLetters,
		}

		continent := classifyRegionByContinent(region)
		continentMap[continent] = append(continentMap[continent], regionInfo)
	}

	// Sort regions within each continent
	for continent := range continentMap {
		sort.Slice(continentMap[continent], func(i, j int) bool {
			return continentMap[continent][i].Name < continentMap[continent][j].Name
		})
	}

	// Select top zones (first available zone in major regions or use custom overrides)
	topZones := selectTopZones(regionMap, customTopZones)

	return Config{
		AllRegions: continentMap,
		TopZones:   topZones,
	}
}

// classifyRegionByContinent provides a rough continent classification
func classifyRegionByContinent(region string) string {
	switch {
	case strings.HasPrefix(region, "us-") || strings.HasPrefix(region, "northamerica-"):
		return "North America"
	case strings.HasPrefix(region, "europe-"):
		return "Europe"
	case strings.HasPrefix(region, "asia-"):
		return "Asia"
	case strings.HasPrefix(region, "australia-"):
		return "Australia"
	case strings.HasPrefix(region, "southamerica-"):
		return "South America"
	case strings.HasPrefix(region, "africa-"):
		return "Africa"
	case strings.HasPrefix(region, "me-"):
		return "Middle East"
	default:
		return "Other"
	}
}

// selectTopZones validates and applies the custom top zones (no defaults since zones are required)
func selectTopZones(regionMap map[string][]string, customTopZones map[string]string) map[string]ZoneConfig {
	topZones := make(map[string]ZoneConfig)

	// Apply and validate custom top zones
	for region, zone := range customTopZones {
		if zones, exists := regionMap[region]; exists {
			// Validate that the specified zone exists for this region
			found := false
			for _, z := range zones {
				if z == zone {
					found = true
					break
				}
			}
			if found {
				topZones[region] = ZoneConfig{Id: zone}
			} else {
				fmt.Printf("Warning: Zone '%s' not found for region '%s', available zones: %v\n", zone, region, zones)
			}
		} else {
			fmt.Printf("Warning: Region '%s' not found in available regions\n", region)
		}
	}

	return topZones
}

// generateFileFromTemplate generates the gcpzones.go file using the template
func generateFileFromTemplate(config Config) error {
	// Read the template file (relative to current directory)
	templatePath := "templates/gcpzones.go.template"
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Create custom template functions
	funcMap := template.FuncMap{
		"join": func(zones []string, sep string) string {
			// Quote each zone and join with separator
			quoted := make([]string, len(zones))
			for i, zone := range zones {
				quoted[i] = fmt.Sprintf(`"%s"`, zone)
			}
			return strings.Join(quoted, sep)
		},
	}

	// Parse the template
	tmpl, err := template.New("gcpzones").Funcs(funcMap).Parse(string(templateContent))
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
	outputPath := filepath.Join(outputDir, "gcpzones.go")
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
	var customTopZones map[string]string

	if *topZonesFlag == "" {
		// Use default hardcoded regions (8 regions for global coverage)
		customTopZones = map[string]string{
			"us-central1":          "a",
			"europe-north1":        "a",
			"asia-northeast1":      "a",
			"australia-southeast2": "a",
			"southamerica-east1":   "a",
			"africa-south1":        "a",
			"me-west1":             "a",
			"asia-south2":          "a",
		}
		fmt.Printf("Using default 8 top zones: %v\n", customTopZones)
	} else {
		// Parse the custom top zones
		customTopZones = parseTopZones(*topZonesFlag)

		// Validate that exactly 2, 4 or 8 top regions are provided
		if len(customTopZones) != 2 && len(customTopZones) != 4 && len(customTopZones) != 8 {
			log.Fatalf("Error: You must provide exactly 2, 4 or 8 top zones, but you provided %d zones.\nProvided zones: %v", len(customTopZones), customTopZones)
		}

		fmt.Printf("Using %d custom top zones: %v\n", len(customTopZones), customTopZones)
	}

	err := GenerateGCPZonesFile(customTopZones)
	if err != nil {
		log.Fatalf("Error generating GCP zones file: %v", err)
	}
}
