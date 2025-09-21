package kubeflake

import "time"

var (
	defaultEpochTime    = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	defaultTimeUnit     = 10 * time.Millisecond
	defaultBitsCluster  = 3
	defaultBitsMachine  = 13
	defaultBitsSequence = 9
)

// Settings configures Kubeflake:
//
// BitsSequence is the bit length of a sequence number.
// If BitsSequence is 0, the default bit length is used, which is 8.
// If BitsSequence is 31 or more, an error is returned.
//
// BitsMachine is the bit length of a machine ID.
// If BitsMachine is 0, the default bit length is used, which is 16.
// If BitsMachine is 17 or more, an error is returned.
//
// BitsCluster is the bit length of a cluster ID.
// If BitsCluster is 0, the default bit length is used, which is 8.
// If BitsCluster is 9 or more (more than 256 clusters), an error is returned.
//
// TimeUnit is the time unit of Kubeflake.
// If TimeUnit is 0, the default time unit is used, which is 10 msec.
// TimeUnit must be 1 msec or longer.
//
// Base is the base encoder used to generate the unique ID from the internal int64.
// By default Base62 will be used.
//
// StartTime is the time since which the Kubeflake time is defined as the elapsed time.
// If StartTime is 0, the start time of the Kubeflake instance is set to "2025-01-01 00:00:00 +0000 UTC".
// StartTime must be before the current time.
//
// MachineID returns the unique ID of a Kubeflake instance.
// If MachineID returns an error, the instance will not be created.
//
// The bit length of time is calculated by 63 - BitsCluster - BitsMachine - BitsSequence.
// If it is less than 32, an error is returned.
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

func DefaultSettings() Settings {
	return Settings{
		BitsSequence: defaultBitsSequence,
		BitsCluster:  defaultBitsCluster,
		BitsMachine:  defaultBitsMachine,
		TimeUnit:     defaultTimeUnit,
		Base:         Base62Converter{},
		EpochTime:    defaultEpochTime,
	}
}
