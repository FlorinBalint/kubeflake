package cloud

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	internal "github.com/FlorinBalint/kubeflake/internal/cloud"
)

type Provider int

const (
	GCPProvider Provider = iota
	AWSProvider
	AzureProvider
	DetectProvider
)

// Additional errors for zone discovery.
var (
	ErrFailedToDetectProvider = errors.New("failed to detect cloud provider")
	// Additional errors for GCP Zone discovery.
	ErrGCPZoneNotFound        = errors.New("gcp zone not found")
	ErrGCPMetadataUnavailable = errors.New("gcp metadata server unavailable")

	// Additional errors for AWS region discovery.
	ErrAWSRegionNotFound      = errors.New("aws region not found")
	ErrAWSMetadataUnavailable = errors.New("aws metadata server unavailable")
)

// gcpZone returns the GCP zone for the current pod's node.
// It checks env overrides (GCP_ZONE, ZONE), then queries the metadata server:
//
//	http://metadata.google.internal/computeMetadata/v1/instance/zone
//
// Requires header: Metadata-Flavor: Google
func gcpZone(ctx context.Context) (string, error) {
	// Env overrides (useful in tests or non-GCP environments)
	if z := strings.TrimSpace(os.Getenv("GCP_ZONE")); z != "" {
		return z, nil
	}
	if z := strings.TrimSpace(os.Getenv("ZONE")); z != "" {
		return z, nil
	}

	// Metadata host override per GCE conventions
	base := "http://metadata.google.internal"
	if h := strings.TrimSpace(os.Getenv("GCE_METADATA_HOST")); h != "" {
		if strings.HasPrefix(h, "http://") || strings.HasPrefix(h, "https://") {
			base = h
		} else {
			base = "http://" + h
		}
	}

	url := base + "/computeMetadata/v1/instance/zone"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
	}
	req.Header.Set("Metadata-Flavor", "Google")

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", ErrGCPMetadataUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ErrGCPMetadataUnavailable
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", ErrGCPMetadataUnavailable
	}
	s := strings.TrimSpace(string(body))
	if s == "" {
		return "", ErrGCPZoneNotFound
	}

	// Response format: projects/<num>/zones/<zone>
	if i := strings.LastIndexByte(s, '/'); i >= 0 && i+1 < len(s) {
		s = s[i+1:]
	}
	if s == "" {
		return "", ErrGCPZoneNotFound
	}
	return s, nil
}

func gcpZoneId(ctx context.Context) (int, error) {
	zone, err := gcpZone(ctx)
	if err != nil {
		return -1, err
	}
	if i, ok := internal.GCPZoneIndex(zone); ok {
		return i, nil
	}
	return -1, ErrGCPZoneNotFound
}

// awsRegion returns the AWS region for the current EC2 instance.
// It checks env overrides (AWS_REGION, AWS_DEFAULT_REGION), then queries the metadata server:
//
//	http://169.254.169.254/latest/meta-data/placement/region
//
// Uses IMDSv2 with token-based authentication for security.
func awsRegion(ctx context.Context) (string, error) {
	// Env overrides (useful in tests or non-AWS environments)
	if r := strings.TrimSpace(os.Getenv("AWS_REGION")); r != "" {
		return r, nil
	}
	if r := strings.TrimSpace(os.Getenv("AWS_DEFAULT_REGION")); r != "" {
		return r, nil
	}

	// AWS EC2 Instance Metadata Service v2 (IMDSv2)
	base := "http://169.254.169.254"
	tokenURL := base + "/latest/api/token"
	regionURL := base + "/latest/meta-data/placement/region"

	// First, get a session token (IMDSv2 requirement)
	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPut, tokenURL, nil)
	if err != nil {
		return "", err
	}
	tokenReq.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600") // 6 hours

	client := &http.Client{Timeout: 2 * time.Second}
	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		return "", ErrAWSMetadataUnavailable
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return "", ErrAWSMetadataUnavailable
	}

	token, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return "", ErrAWSMetadataUnavailable
	}

	// Now get the region using the token
	regionReq, err := http.NewRequestWithContext(ctx, http.MethodGet, regionURL, nil)
	if err != nil {
		return "", err
	}
	regionReq.Header.Set("X-aws-ec2-metadata-token", string(token))

	regionResp, err := client.Do(regionReq)
	if err != nil {
		return "", ErrAWSMetadataUnavailable
	}
	defer regionResp.Body.Close()

	if regionResp.StatusCode != http.StatusOK {
		return "", ErrAWSMetadataUnavailable
	}

	body, err := io.ReadAll(regionResp.Body)
	if err != nil {
		return "", ErrAWSMetadataUnavailable
	}

	region := strings.TrimSpace(string(body))
	if region == "" {
		return "", ErrAWSRegionNotFound
	}
	return region, nil
}

func awsRegionId(ctx context.Context) (int, error) {
	region, err := awsRegion(ctx)
	if err != nil {
		return -1, err
	}
	if i, ok := internal.AWSRegionIndex(region); ok {
		return i, nil
	}
	return -1, ErrAWSRegionNotFound
}

func detectProvider(ctx context.Context) (Provider, error) {
	// TODO: implement platform detection
	return GCPProvider, nil
}

// AvailabilityZoneId returns the availability zone ID for the given provider.
// For GCP, this returns the zone index. For AWS, this returns the region index.
func AvailabilityZoneId(provider Provider) (int, error) {
	switch provider {
	case GCPProvider:
		return gcpZoneId(context.Background())
	case AWSProvider:
		return awsRegionId(context.Background())
	case DetectProvider:
		detected, err := detectProvider(context.Background())
		if err != nil {
			return -1, err
		}
		return AvailabilityZoneId(detected)
	default:
		// TODO: implement for Azure
		return -1, fmt.Errorf("function not implemented for provider: %v", provider)
	}
}
