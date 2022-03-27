package tasks

import (
	"context"
	"log"
	"math/rand"
	"testutil/cli"
	"time"
)

type RestartStrategy interface {
	Init(config *Tester, ids []string)
	Run(ctx context.Context)
}

type NoRestartStrategy struct{}

func (n *NoRestartStrategy) Init(config *Tester, ids []string) {
	cli.DockerStart(ids...)
}

func (n *NoRestartStrategy) Run(ctx context.Context) {
	// nothing to do
}

type RandomRestartsStrategy struct {
	TimeToWork   time.Duration
	RestartEvery time.Duration

	GlobalRestart time.Duration
	NextGlobal    time.Time

	lastRestarted []time.Time
	ids           []string
}

func (r *RandomRestartsStrategy) Init(config *Tester, ids []string) {
	// don't start everything, restarts should handle it
	r.NextGlobal = time.Now().Add(r.GlobalRestart)
	r.ids = ids
	r.lastRestarted = make([]time.Time, len(ids))
}

func (r *RandomRestartsStrategy) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if time.Now().After(r.NextGlobal) {
			err := cli.DockerRestart(r.ids...)
			if err != nil {
				log.Printf("WARN error restarting: %v", err)
			}
			r.NextGlobal = time.Now().Add(r.GlobalRestart)
		}

		index := -1
		for i := 0; i < 1000; i++ {
			index = rand.Intn(len(r.ids))
			if time.Since(r.lastRestarted[index]) > r.RestartEvery {
				break
			}
		}
		if index == -1 {
			// no one to restart
		} else {
			err := cli.DockerRestart(r.ids[index])
			if err != nil {
				log.Printf("WARN error restarting: %v", err)
			}
			r.lastRestarted[index] = time.Now()
		}

		time.Sleep(r.RestartEvery)
	}
}

type EpochStrategy struct {
	EpochTime time.Duration
	Nodes     int
	ids       []string
}

func (e *EpochStrategy) Init(config *Tester, ids []string) {
	e.ids = ids
}

func (e *EpochStrategy) Run(ctx context.Context) {
	defer cli.DockerKill(e.ids...)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ids := append([]string{}, e.ids...)
		rand.Shuffle(len(ids), func(i, j int) {
			ids[i], ids[j] = ids[j], ids[i]
		})
		ids = ids[:e.Nodes]

		err := cli.DockerRestart(ids...)
		if err != nil {
			log.Printf("WARN error restarting: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(e.EpochTime):
		}

		_ = cli.DockerKill(e.ids...)
	}
}
