package kubernetes

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

// Errors returned by StatefulSetPod.
var (
	ErrPodNameNotFound = errors.New("statefulset pod name not found from environment or hostname")
	ErrOrdinalNotFound = errors.New("ordinal suffix not found or not numeric in pod name")
)

// PodName retrieves the name of the current pod from the environment or hostname.
func PodName() (string, error) {
	// Preferred: Downward API provided POD_NAME
	if name := os.Getenv("POD_NAME"); name != "" {
		return name, nil
	}
	// Kubernetes typically sets HOSTNAME to the pod name
	if name := os.Getenv("HOSTNAME"); name != "" {
		return name, nil
	}

	// Fallback to system hostname
	hn, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hn, nil
}

// A function that retrieves the Machine ID from the StatefulSet's ordinal index.
func StatefulSetPodId() (int, error) {
	// Get the pod name
	podName, err := PodName()
	if err != nil {
		return 0, ErrPodNameNotFound
	}
	// Extract the ordinal index from the pod name
	idx := strings.LastIndex(podName, "-")
	if idx < 0 || idx == len(podName)-1 {
		return 0, ErrOrdinalNotFound
	}
	suffix := podName[idx+1:]
	n, convErr := strconv.Atoi(suffix)
	if convErr != nil {
		return 0, ErrOrdinalNotFound
	}
	return n, nil
}
