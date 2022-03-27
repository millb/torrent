package tasks

import (
	"math/rand"
	"testutil"
	"testutil/cli"
	"time"
)

func RestartsDiscovery() bool {
	subnet := "172.20.17.0/24"
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

		if !RestartsDiscoveryRun(nodesCount, networkID, subnet, dataSize, partSize, strategy1) {
			return false
		}
		if !RestartsDiscoveryRun(nodesCount, networkID, subnet, dataSize, partSize, strategy2) {
			return false
		}
	}

	return true
}

func RestartsDiscoveryRun(nodesCount int, network string, subnet string, dataSize int, partSize int, strategy RestartStrategy) bool {
	var nodes []Node
	for i := 0; i < nodesCount; i++ {
		var prev int
		if i > 0 {
			prev = rand.Intn(i)
		}

		nodes = append(nodes, Node{
			HasFullData:     false,
			HasAllPeers:     false,
			HasFileInfo:     false,
			AllowExtraTrash: false,
			HasSinglePeer:   true,
			SinglePeerIndex: prev,
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
