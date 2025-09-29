package kubeflake

import (
	"errors"
	"time"

	"github.com/FlorinBalint/kubeflake/pkg/cloud"
	"github.com/FlorinBalint/kubeflake/pkg/kubernetes"
)

const (
	DefaultTimeUnit     = 10 * time.Millisecond
	DefaultBitsCluster  = 3
	DefaultBitsMachine  = 13
	DefaultBitsSequence = 9
	// Bit lengths constraints
	MinTimeBits     = 30
	MinSequenceBits = 8
	MaxSequenceBits = 30
	MinClusterBits  = 2
	MaxClusterBits  = 8
	MaxMachineBits  = 16
	MinMachineBits  = 3
)

var defaultEpochTime = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

var (
	ErrInvalidBitsTime      = errors.New("bit length for time must be 32 or more")
	ErrInvalidBitsSequence  = errors.New("invalid bit length for sequence number")
	ErrInvalidBitsMachineID = errors.New("invalid bit length for machine id")
	ErrInvalidBitsClusterID = errors.New("invalid bit length for cluster id")
	ErrInvalidTimeUnit      = errors.New("invalid time unit")
	ErrInvalidSequence      = errors.New("invalid sequence number")
	ErrInvalidMachineID     = errors.New("invalid machine id")
	ErrInvalidClusterID     = errors.New("invalid cluster id")
	ErrStartTimeAhead       = errors.New("start time is ahead")
	ErrOverTimeLimit        = errors.New("over the time limit")
)

// Settings configures Kubeflake:
//
// BitsSequence is the bit length of a sequence number.
// If BitsSequence is 0, the default bit length is used, which is 8.
// A BitsSequence of 31 or more is considered invalid.
//
// BitsCluster is the bit length of a cluster ID.
// A BitsCluster of 9 or more (more than 256 clusters) is considered invalid.
// ClusterID returns the unique ID of a cluster.
// The ClusterID function returns the unique ID of a cluster.
// ClusterID must return a value between 0 and 2^BitsCluster - 1.
//
// BitsMachine is the bit length of a machine ID.
// A BitsMachine of 17 or more is considered invalid.
// The MachineID function returns the unique ID of a Kubeflake instance within a cluster.
// MachineID must return a value between 0 and 2^BitsMachine - 1.
//
// Base is the base encoder used to generate the unique ID from the internal int64.
// By default Base62 will be used.
//
// StartTime is the time since which the Kubeflake time is defined as the elapsed time.
// StartTime must be before the current time.
//
// TimeUnit is the time unit of Kubeflake.
// TimeUnit must be 1 msec or longer.
//
// The bit length of time is calculated by 64 - BitsCluster - BitsMachine - BitsSequence.
// If it is less than 30, an error is returned (12.4 years with 1 msec time unit).
// TODO: Consider allowing fewer bits if timeunit is 10 msec or more.
type Settings struct {
	BitsSequence int
	BitsCluster  int
	BitsMachine  int

	TimeUnit  time.Duration
	Base      BaseConverter
	EpochTime time.Time
	ClusterId func() (int, error)
	MachineId func() (int, error)
}

func (s Settings) Validate() error {
	// Validate settings
	if s.BitsSequence < MinSequenceBits || s.BitsSequence > MaxSequenceBits {
		return ErrInvalidBitsSequence
	}
	if s.BitsMachine < MinMachineBits || s.BitsMachine > MaxMachineBits {
		return ErrInvalidBitsMachineID
	}
	if s.BitsCluster < MinClusterBits || s.BitsCluster > MaxClusterBits {
		return ErrInvalidBitsClusterID
	}
	if s.TimeUnit < 0 || (s.TimeUnit > 0 && s.TimeUnit < time.Millisecond) {
		return ErrInvalidTimeUnit
	}
	if s.EpochTime.After(time.Now()) {
		return ErrStartTimeAhead
	}
	bitsTime := 64 - s.BitsCluster - s.BitsMachine - s.BitsSequence
	if bitsTime < MinTimeBits {
		return ErrInvalidBitsTime
	}
	return nil
}

func detectAZId() (int, error) {
	return cloud.AvailabilityZoneId(cloud.DetectProvider)
}

func DefaultSettings() Settings {
	return Settings{
		BitsSequence: DefaultBitsSequence,
		BitsCluster:  DefaultBitsCluster,
		BitsMachine:  DefaultBitsMachine,
		TimeUnit:     DefaultTimeUnit,
		Base:         Base62Converter{},
		EpochTime:    defaultEpochTime,
		MachineId:    kubernetes.StatefulSetPodId,
		ClusterId:    detectAZId,
	}
}
