package tasks

import (
	"math/rand"
	"testutil"
	"testutil/cli"
	"time"
)

func SelfDiscovery() bool {
	subnet := "172.20.18.0/24"
	networkID := cli.DockerNetworkCreate(subnet, testutil.Network)
	defer cli.DockerNetworkRm(testutil.Network)

	for _, nodesCount := range []int{4, 8, 14} {
		const (
			megabyte = 1024 * 1024
			kilobyte = 1024
		)
		dataSize := megabyte + rand.Intn(megabyte*16)
		partSize := (64 + rand.Intn(64)) * kilobyte
		strategy1 := &RandomRestartsStrategy{
			TimeToWork:    time.Second * 3,
			RestartEvery:  time.Second,
			GlobalRestart: time.Second * 10,
		}
		strategy2 := &EpochStrategy{
			EpochTime: time.Second * 3,
			Nodes:     nodesCount / 2,
		}

		if !SelfDiscoveryRun(nodesCount, networkID, subnet, dataSize, partSize, strategy1) {
			return false
		}
		if !SelfDiscoveryRun(nodesCount, networkID, subnet, dataSize, partSize, strategy2) {
			return false
		}
	}

	return true
}

func SelfDiscoveryRun(nodesCount int, network string, subnet string, dataSize int, partSize int, strategy RestartStrategy) bool {
	var nodes []Node
	for i := 0; i < nodesCount; i++ {
		nodes = append(nodes, Node{
			HasFullData:     false,
			HasAllPeers:     false,
			HasFileInfo:     false,
			AllowExtraTrash: false,
		})
	}

	nodes[rand.Intn(nodesCount)].HasFileInfo = true

	config := &Tester{
		NodesCount:      nodesCount,
		Subnet:          subnet,
		Nodes:           nodes,
		Network:         network,
		DataSize:        dataSize,
		PartSize:        partSize,
		Timeout:         time.Second * 120,
		RandomParts:     true,
		RestartStrategy: strategy,
	}
	return InitAndRun(config)
}