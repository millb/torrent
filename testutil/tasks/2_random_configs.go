package tasks

import (
	"math/rand"
	"testutil"
	"testutil/cli"
	"time"
)

func SimpleRandomConfigs() bool {
	subnet := "172.20.234.0/24"
	networkID := cli.DockerNetworkCreate(subnet, testutil.Network)
	defer cli.DockerNetworkRm(testutil.Network)

	for _, nodesCount := range []int{2, 3, 5, 10} {
		if !SimpleRandomConfig(nodesCount, networkID, subnet) {
			return false
		}
	}

	return true
}

func SimpleRandomConfig(nodesCount int, network string, subnet string) bool {
	var nodes []Node
	for i := 0; i < nodesCount; i++ {
		nodes = append(nodes, Node{
			HasFullData: false,
			HasAllPeers: true,
			HasFileInfo: true,
		})
	}
	nodes[len(nodes)-1].HasFullData = true

	config := &Tester{
		NodesCount: nodesCount,
		Subnet:     subnet,
		Nodes:      nodes,
		Network:    network,
		DataSize:   1024 * (1 + rand.Intn(1024)),
		PartSize:   128 * (1 + rand.Intn(32)),
		Timeout:    time.Second * 30,
	}
	return InitAndRun(config)
}
