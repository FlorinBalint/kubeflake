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

// AWSRegion represents an AWS region from the AWS CLI output
type AWSRegion struct {
	Endpoint    string `json:"Endpoint"`
	RegionName  string `json:"RegionName"`
	OptInStatus string `json:"OptInStatus"`
}

// AWSRegionsOutput represents the output from describe-regions
type AWSRegionsOutput struct {
	Regions []AWSRegion `json:"Regions"`
}

// Config represents the template configuration
type Config struct {
	AllRegions []string
	TopRegions []string
}

// TemplateData represents the data passed to the template
type TemplateData struct {
	Config Config
}

// parseTopRegions parses the top regions override flag format: "region1,region2,region3"
func parseTopRegions(topRegionsFlag string) []string {
	if topRegionsFlag == "" {
		return []string{}
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

// GenerateAWSRegionsFile runs AWS CLI commands and generates awsregions.go file
func GenerateAWSRegionsFile(customTopRegions []string) error {
	// Get AWS regions
	regions, err := getAWSRegions()
	if err != nil {
		return fmt.Errorf("failed to get AWS regions: %w", err)
	}

	// Process regions into config
	config := processRegionsIntoConfig(regions, customTopRegions)

	// Generate the file from template
	err = generateFileFromTemplate(config)
	if err != nil {
		return fmt.Errorf("failed to generate file from template: %w", err)
	}

	fmt.Println("Successfully generated awsregions.go")
	return nil
}

// getAWSRegions runs AWS CLI to get all regions
func getAWSRegions() ([]AWSRegion, error) {
	cmd := exec.Command("aws", "ec2", "describe-regions", "--all-regions", "--output=json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run aws ec2 describe-regions command: %w", err)
	}

	var regionsOutput AWSRegionsOutput
	err = json.Unmarshal(output, &regionsOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w", err)
	}

	return regionsOutput.Regions, nil
}

// processRegionsIntoConfig converts regions into the config structure expected by the template
func processRegionsIntoConfig(regions []AWSRegion, customTopRegions []string) Config {
	allRegions := make([]string, 0, len(regions))

	// Collect all opted-in regions
	for _, region := range regions {
		// Include regions that are opted-in or opt-in-not-required
		if region.OptInStatus == "opted-in" || region.OptInStatus == "opt-in-not-required" {
			allRegions = append(allRegions, region.RegionName)
		}
	}

	sort.Strings(allRegions)

	// Validate custom top regions
	validTopRegions := make([]string, 0, len(customTopRegions))
	regionSet := make(map[string]bool)
	for _, r := range allRegions {
		regionSet[r] = true
	}

	for _, region := range customTopRegions {
		if regionSet[region] {
			validTopRegions = append(validTopRegions, region)
		} else {
			fmt.Printf("Warning: Region '%s' not found in available regions\n", region)
		}
	}

	return Config{
		AllRegions: allRegions,
		TopRegions: validTopRegions,
	}
}

// generateFileFromTemplate generates the awsregions.go file using the template
func generateFileFromTemplate(config Config) error {
	// Read the template file (relative to current directory)
	templatePath := "templates/awsregions.go.template"
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Create custom template functions
	funcMap := template.FuncMap{
		"join": func(items []string, sep string) string {
			// Quote each item and join with separator
			quoted := make([]string, len(items))
			for i, item := range items {
				quoted[i] = fmt.Sprintf(`"%s"`, item)
			}
			return strings.Join(quoted, sep)
		},
	}

	// Parse the template
	tmpl, err := template.New("awsregions").Funcs(funcMap).Parse(string(templateContent))
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
	outputPath := filepath.Join(outputDir, "awsregions.go")
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
	var (
		topRegionsFlag string
	)

	flag.StringVar(&topRegionsFlag, "top-regions", "", "Override top regions in format 'region1,region2,region3' (exactly 2, 4 or 8 regions, e.g., 'us-east-1,eu-west-1,ap-southeast-1,us-west-2'). If not provided, uses default regions.")
	flag.Parse()

	var customTopRegions []string

	if topRegionsFlag == "" {
		// Use default hardcoded regions (8 regions for global coverage)
		customTopRegions = []string{
			"ap-southeast-2", // Sydney
			"eu-west-2",      // London
			"us-west-1",      // North California
			"ap-east-1",      // Hong Kong
			"af-south-1",     // South Africa, Cape Town
			"sa-east-1",      // São Paulo
			"me-central-1",   // Bahrain
			"ca-central-1",   // Canada Central
		}
		fmt.Printf("Using default 8 top regions: %v\n", customTopRegions)
		fmt.Println("Note: Ensure you have AWS credentials configured and the AWS CLI is accessible.")
	} else {
		// Parse the custom top regions
		customTopRegions = parseTopRegions(topRegionsFlag)

		// Validate that exactly 2, 4 or 8 top regions are provided
		if len(customTopRegions) != 2 && len(customTopRegions) != 4 && len(customTopRegions) != 8 {
			log.Fatalf("Error: You must provide exactly 2, 4 or 8 top regions, but you provided %d regions.\nProvided regions: %v", len(customTopRegions), customTopRegions)
		}

		fmt.Printf("Using %d custom top regions: %v\n", len(customTopRegions), customTopRegions)
	}

	err := GenerateAWSRegionsFile(customTopRegions)
	if err != nil {
		log.Fatalf("Error generating AWS regions file: %v", err)
	}
}
