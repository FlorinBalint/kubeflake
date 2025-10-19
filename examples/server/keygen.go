package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/FlorinBalint/kubeflake/pkg/cloud"
	"github.com/FlorinBalint/kubeflake/pkg/kubernetes"
	kf "github.com/FlorinBalint/kubeflake/v1"
)

var (
	listenAddr   = flag.String("address", ":8083", "HTTP listen address")
	bitsMachine  = flag.Int("bits.machine", 6, "Number of bits for machine ID")
	bitsSequence = flag.Int("bits.sequence", 11, "Number of bits for sequence ID")
	bitsCluster  = flag.Int("bits.cluster", 7, "Number of bits for cluster ID")
)

type keygenHandler struct {
	kubeFlake *kf.Kubeflake
}

func gcpZoneId() (int, error) {
	return cloud.AvailabilityZoneId(cloud.GCPProvider)
}

func newHandler() (*keygenHandler, error) {
	flakeAlgo, err := kf.New(
		kf.WithMachineBits(*bitsMachine),
		kf.WithClusterBits(*bitsCluster),
		kf.WithSequenceBits(*bitsSequence),
		kf.WithTimeUnit(10*time.Millisecond),
		kf.WithBase62Keys(),
		kf.WithMachineIdFn(kubernetes.StatefulSetPodId),
		kf.WithClusterIdFn(gcpZoneId),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create Kubeflake: %w", err)
	}

	return &keygenHandler{
		kubeFlake: flakeAlgo,
	}, nil
}

func (h *keygenHandler) generateKey(w http.ResponseWriter, r *http.Request) {
	key, err := h.kubeFlake.NextKey()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate key: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err = w.Write([]byte(key))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *keygenHandler) generateID(w http.ResponseWriter, r *http.Request) {
	id, err := h.kubeFlake.NextID()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate ID: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err = w.Write([]byte(fmt.Sprintf("%d", id)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *keygenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/health":
		w.WriteHeader(http.StatusOK)
	case "/generate/v1":
		h.generateKey(w, r)
	case "/generate/v1/id":
		h.generateID(w, r)
	default:
		http.NotFound(w, r)
	}
}

func main() {
	handler, err := newHandler()
	if err != nil {
		fmt.Println("Error creating handler:", err)
		return
	}

	http.Handle("/", handler)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		panic(err)
	}
}
