package tasks

import (
	"math/rand"
	"testutil"
	"testutil/cli"
	"time"
)

func SimpleExtraCases() bool {
	subnet := "172.20.14.0/24"
	networkID := cli.DockerNetworkCreate(subnet, testutil.Network)
	defer cli.DockerNetworkRm(testutil.Network)

	if !SimpleFixedSize(7, 1, networkID, subnet, 0, 1) {
		return false
	}

	for _, nodesCount := range []int{3, 9, 30, 60} {
		const (
			gigabyte = 1024 * 1024 * 1024
			kilobyte = 1024
		)
		dataSize := gigabyte / (nodesCount + 1)

		nodesHaveData := nodesCount / 8
		if nodesHaveData < 1 {
			nodesHaveData = 1
		}

		if !SimpleFixedSize(nodesCount, nodesHaveData, networkID, subnet, dataSize, 64*kilobyte) {
			return false
		}
	}

	return true
}

func SimpleFixedSize(nodesCount int, nodesHaveData int, network string, subnet string, dataSize int, partSize int) bool {
	var nodes []Node
	for i := 0; i < nodesCount; i++ {
		haveData := i < nodesHaveData
		nodes = append(nodes, Node{
			HasFullData:     haveData,
			HasAllPeers:     haveData || (rand.Intn(2) == 1),
			HasFileInfo:     true,
			AllowExtraTrash: !haveData,
		})
	}

	config := &Tester{
		NodesCount: nodesCount,
		Subnet:     subnet,
		Nodes:      nodes,
		Network:    network,
		DataSize:   dataSize,
		PartSize:   partSize,
		Timeout:    time.Second * 60,
	}
	return InitAndRun(config)
}
