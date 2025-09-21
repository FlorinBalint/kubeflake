package kubeflake

import (
	"time"

	internal "github.com/FlorinBalint/kubeflake/internal/kubeflake"
)

// GeneratorOptions defines functional options for Kubeflake generator
type GeneratorOptions interface {
	apply(*settings)
}
type optionFunc func(*settings)

func (f optionFunc) apply(s *settings) {
	f(s)
}

// WithSequenceBits sets the number of bits for sequence
func WithSequenceBits(bits int) GeneratorOptions {
	return optionFunc(func(s *settings) {
		s.BitsSequence = bits
	})
}

// WithClusterBits sets the number of bits for cluster ID
func WithClusterBits(bits int) GeneratorOptions {
	return optionFunc(func(s *settings) {
		s.BitsCluster = bits
	})
}

// WithMachineBits sets the number of bits for machine ID
func WithMachineBits(bits int) GeneratorOptions {
	return optionFunc(func(s *settings) {
		s.BitsMachine = bits
	})
}

// WithTimeUnit sets the time unit
func WithTimeUnit(unit time.Duration) GeneratorOptions {
	return optionFunc(func(s *settings) {
		s.TimeUnit = unit
	})
}

// WithBase62Keys converts ids using base62
func WithBase62Keys() GeneratorOptions {
	return optionFunc(func(s *settings) {
		s.Base = internal.Base62Converter{}
	})
}

// WithBase64Keys converts ids using base64
func WithBase64Keys() GeneratorOptions {
	return optionFunc(func(s *settings) {
		s.Base = internal.Base64Converter{}
	})
}

// WithEpoch sets the epoch time
func WithEpoch(t time.Time) GeneratorOptions {
	return optionFunc(func(s *settings) {
		s.EpochTime = t
	})
}

// WithClusterId sets the cluster ID function
func WithClusterIdFn(fn func() (int, error)) GeneratorOptions {
	return optionFunc(func(s *settings) {
		s.ClusterId = fn
	})
}

// WithMachineId sets the machine ID function
func WithMachineIdFn(fn func() (int, error)) GeneratorOptions {
	return optionFunc(func(s *settings) {
		s.MachineId = fn
	})
}
