package tasks

import (
	"math/rand"
	"testutil"
	"testutil/cli"
	"time"
)

func ManySource() bool {
	subnet := "172.20.15.0/24"
	networkID := cli.DockerNetworkCreate(subnet, testutil.Network)
	defer cli.DockerNetworkRm(testutil.Network)

	for _, nodesCount := range []int{4, 8, 16} {
		const (
			megabyte = 1024 * 1024
			kilobyte = 1024
		)
		dataSize := megabyte + rand.Intn(megabyte*16)
		partSize := (64 + rand.Intn(64)) * kilobyte

		if !ManySourceRun(nodesCount, networkID, subnet, dataSize, partSize) {
			return false
		}
	}

	return true
}

func ManySourceRun(nodesCount int, network string, subnet string, dataSize int, partSize int) bool {
	var nodes []Node
	for i := 0; i < nodesCount; i++ {
		nodes = append(nodes, Node{
			HasFullData:     false,
			HasAllPeers:     true,
			HasFileInfo:     true,
			AllowExtraTrash: false,
		})
	}

	config := &Tester{
		NodesCount:  nodesCount,
		Subnet:      subnet,
		Nodes:       nodes,
		Network:     network,
		DataSize:    dataSize,
		PartSize:    partSize,
		Timeout:     time.Second * 30,
		RandomParts: true,
	}
	return InitAndRun(config)
}
