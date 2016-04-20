package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/docker/swarm-v2/api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Config holds the benchmarking configuration.
type Config struct {
	Count   int64
	Manager string
	IP      string
	Port    int
	Unit    time.Duration
}

// Benchmark represents a benchmark session.
type Benchmark struct {
	cfg       *Config
	collector *Collector
}

// NewBenchmark creates a new benchmark session with the given configuration.
func NewBenchmark(cfg *Config) *Benchmark {
	return &Benchmark{
		cfg:       cfg,
		collector: NewCollector(),
	}
}

// Run starts the benchmark session and waits for it to be completed.
func (b *Benchmark) Run() error {
	fmt.Printf("Listening for incoming connections at %s:%d\n", b.cfg.IP, b.cfg.Port)
	if err := b.collector.Listen(b.cfg.Port); err != nil {
		return err
	}
	j, err := b.launch()
	if err != nil {
		return err
	}
	fmt.Printf("Service %s launched (%d instances)\n", j.ID, b.cfg.Count)

	// Periodically print stats.
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(5 * time.Second):
				fmt.Printf("\n%s: Progression report\n", time.Now())
				b.collector.Stats(os.Stdout, time.Second)
			case <-doneCh:
				return
			}
		}
	}()

	fmt.Println("Collecting metrics...")
	b.collector.Collect(b.cfg.Count)
	doneCh <- struct{}{}

	fmt.Printf("\n%s: Benchmark completed\n", time.Now())
	b.collector.Stats(os.Stdout, time.Second)

	return nil
}

func (b *Benchmark) spec() *api.ServiceSpec {
	return &api.ServiceSpec{
		Annotations: api.Annotations{
			Name: "benchmark",
		},
		Template: &api.TaskSpec{
			Runtime: &api.TaskSpec_Container{
				Container: &api.Container{
					Image: &api.Image{
						Reference: "alpine:latest",
					},
					Command: []string{"nc", b.cfg.IP, strconv.Itoa(b.cfg.Port)},
				},
			},
		},
		Instances: b.cfg.Count,
	}
}

func (b *Benchmark) launch() (*api.Service, error) {
	conn, err := grpc.Dial(b.cfg.Manager, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := api.NewClusterClient(conn)
	r, err := client.CreateService(context.Background(), &api.CreateServiceRequest{
		Spec: b.spec(),
	})
	if err != nil {
		return nil, err
	}
	return r.Service, nil
}
