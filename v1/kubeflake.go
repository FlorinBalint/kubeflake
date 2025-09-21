package kubeflake

import (
	"sync"
	"time"

	"errors"

	internal "github.com/FlorinBalint/kubeflake/internal/kubeflake"
)

type IdParts string
type settings = internal.Settings
type baseConverter = internal.BaseConverter

const (
	Timestamp IdParts = "timestamp"
	Sequence  IdParts = "sequence"
	MachineID IdParts = "machine_id"
	ClusterID IdParts = "cluster_id"
)

var (
	errInvalidSequence  = errors.New("invalid sequence number")
	errInvalidMachineID = errors.New("invalid machine id")
	errInvalidClusterID = errors.New("invalid cluster id")
	errOverTimeLimit    = errors.New("over the time limit")
)

type Kubeflake struct {
	mutex     *sync.Mutex
	machineId int
	clusterId int

	bitsTime     int
	bitsCluster  int
	bitsMachine  int
	bitsSequence int
	sequenceMask uint64

	timeUnit    int64
	startTime   uint64
	elapsedTime uint64

	sequence uint64
	base     baseConverter
	nowFunc  func() time.Time
}

// New creates a new Kubeflake with the given options
// If an option is provided, the Kubeflake instance uses the default value for that option.
// The MachineId function and the ClusterId function must be provided.
// TODO: Add default MachineId and ClusterId options
//
// The default settings are:
// - BitsSequence: 8
// - BitsCluster: 8
// - BitsMachine: 16
// - TimeUnit: 10 msec
// - Base: Base62Converter
// - EpochTime: "2025-01-01 00:00:00 +0000 UTC"
func New(opts ...GeneratorOptions) (*Kubeflake, error) {
	s := internal.DefaultSettings()
	for _, opt := range opts {
		opt.apply(&s)
	}
	return newWithSettings(s)
}

// New returns a new Kubeflake configured with the given Settings.
// New returns an error in the following cases:
// - Settings.BitsSequence is less than 0 or greater than 30.
// - Settings.BitsMachineID is less than 0 or greater than 30.
// - Settings.BitsSequence + Settings.BitsMachineID is 32 or more.
// - Settings.TimeUnit is less than 1 msec.
// - Settings.StartTime is ahead of the current time.
// - Settings.MachineID returns an error.
// - Settings.ClusterId returns an error.
func newWithSettings(settings settings) (*Kubeflake, error) {
	// Validate settings
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	k8sFlake := new(Kubeflake)
	k8sFlake.mutex = new(sync.Mutex)
	k8sFlake.nowFunc = time.Now
	k8sFlake.base = settings.Base
	k8sFlake.timeUnit = settings.TimeUnit.Nanoseconds()
	k8sFlake.startTime = k8sFlake.toInternalTime(settings.EpochTime)
	k8sFlake.bitsCluster = settings.BitsCluster
	k8sFlake.bitsMachine = settings.BitsMachine
	k8sFlake.bitsSequence = settings.BitsSequence
	k8sFlake.sequenceMask = uint64(1<<k8sFlake.bitsSequence - 1)
	k8sFlake.bitsTime = 64 - k8sFlake.bitsCluster - k8sFlake.bitsMachine - k8sFlake.bitsSequence

	if cluster, err := settings.ClusterId(); err != nil {
		return nil, err
	} else {
		k8sFlake.clusterId = cluster
	}

	if machine, err := settings.MachineId(); err != nil {
		return nil, err
	} else {
		k8sFlake.machineId = machine
	}

	return k8sFlake, nil
}

func (kf *Kubeflake) toInternalTime(t time.Time) uint64 {
	return uint64(t.UTC().UnixNano() / kf.timeUnit)
}

func (kf *Kubeflake) currentElapsedTime() uint64 {
	return kf.toInternalTime(kf.nowFunc()) - kf.startTime
}

func (kf *Kubeflake) sleep(overtime int64) {
	sleepTime := time.Duration(overtime*kf.timeUnit) -
		time.Duration(kf.nowFunc().UTC().UnixNano()%kf.timeUnit)
	time.Sleep(sleepTime)
}

// NextKey generates a next unique ID as a base-encoded string.
func (kf *Kubeflake) NextKey() (string, error) {
	id, err := kf.NextID()
	if err != nil {
		return "", err
	}
	return kf.base.Encode(id), nil
}

// NextID generates a next unique ID as uint64.
// After the Kubeflake time overflows, NextID returns an error.
func (kf *Kubeflake) NextID() (uint64, error) {

	kf.mutex.Lock()
	defer kf.mutex.Unlock()

	current := kf.currentElapsedTime()
	if kf.elapsedTime < current {
		kf.elapsedTime = current
		kf.sequence = 0
	} else {
		kf.sequence = (kf.sequence + 1) & kf.sequenceMask
		if kf.sequence == 0 {
			kf.elapsedTime++
			overtime := kf.elapsedTime - current
			kf.sleep(int64(overtime))
		}
	}

	return kf.toID()
}

func (kf *Kubeflake) toID() (uint64, error) {
	if kf.elapsedTime >= 1<<kf.bitsTime {
		return 0, errOverTimeLimit
	}

	res := kf.elapsedTime << (kf.bitsSequence + kf.bitsCluster + kf.bitsMachine)
	res |= uint64(kf.sequence) << (kf.bitsMachine + kf.bitsCluster)
	res |= uint64(kf.clusterId) << kf.bitsMachine
	res |= uint64(kf.machineId)
	return res, nil
}

func (kf *Kubeflake) ComposeKey(t time.Time, sequence, machineID, clusterId int) (string, error) {
	id, err := kf.Compose(t, sequence, machineID, clusterId)
	if err != nil {
		return "", err
	}
	return kf.base.Encode(id), nil
}

func (kf *Kubeflake) Compose(t time.Time, sequence, machineID, clusterId int) (uint64, error) {
	internalTime := kf.toInternalTime(t.UTC())
	if internalTime < kf.startTime {
		return 0, internal.ErrStartTimeAhead
	}
	elapsedTime := internalTime - kf.startTime
	if elapsedTime >= 1<<kf.bitsTime {
		return 0, errOverTimeLimit
	}

	if sequence < 0 || sequence >= 1<<kf.bitsSequence {
		return 0, errInvalidSequence
	}

	if clusterId < 0 || clusterId >= 1<<kf.bitsCluster {
		return 0, errInvalidClusterID
	}

	if machineID < 0 || machineID >= 1<<kf.bitsMachine {
		return 0, errInvalidMachineID
	}

	return elapsedTime<<(kf.bitsSequence+kf.bitsMachine+kf.bitsCluster) |
		uint64(sequence)<<(kf.bitsMachine+kf.bitsCluster) |
		uint64(clusterId)<<kf.bitsMachine |
		uint64(machineID), nil
}

func (kf *Kubeflake) DecomposeKey(key string) (map[IdParts]uint64, error) {
	id, err := kf.base.Decode(key)
	if err != nil {
		return nil, err
	}
	return kf.Decompose(id), nil
}

func (kf *Kubeflake) Decompose(id uint64) map[IdParts]uint64 {
	return map[IdParts]uint64{
		Timestamp: kf.timePart(id),
		Sequence:  kf.sequencePart(id),
		MachineID: kf.machinePart(id),
		ClusterID: kf.clusterPart(id),
	}
}

func (kf *Kubeflake) timePart(id uint64) uint64 {
	return uint64(id >> (kf.bitsSequence + kf.bitsCluster + kf.bitsMachine))
}

func (kf *Kubeflake) sequencePart(id uint64) uint64 {
	maskSequence := (1<<kf.bitsSequence - 1) << (kf.bitsMachine + kf.bitsCluster)
	return (id & uint64(maskSequence)) >> (kf.bitsMachine + kf.bitsCluster)
}

func (kf *Kubeflake) clusterPart(id uint64) uint64 {
	maskCluster := (1<<kf.bitsCluster - 1) << kf.bitsMachine
	return (id & uint64(maskCluster)) >> kf.bitsMachine
}

func (kf *Kubeflake) machinePart(id uint64) uint64 {
	maskMachine := uint64(1<<kf.bitsMachine - 1)
	return id & maskMachine
}
