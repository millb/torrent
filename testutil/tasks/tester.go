package tasks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testutil"
	"testutil/cli"
	"time"
)

type Tester struct {
	NodesCount      int
	Subnet          string
	Seed            int64
	Network         string
	Timeout         time.Duration
	Nodes           []Node
	RandomParts     bool
	RestartStrategy RestartStrategy
	SaveLogsPath    string

	DataSize int
	PartSize int

	Gen       *testutil.Generator
	Data      *os.File
	Hashes    []testutil.Hash
	Sha256Sum string
	IPs       []string
}

type Node struct {
	HasFullData     bool
	HasFileInfo     bool
	HasAllPeers     bool
	AllowExtraTrash bool

	HasSinglePeer   bool
	SinglePeerIndex int

	PartsMap map[int]bool

	IP       string
	DockerID string

	Config *testutil.Config
}

func InitAndRun(config *Tester) bool {
	if config.Network == "" {
		config.Network = testutil.Network
	}
	if config.Seed == 0 {
		config.Seed = time.Now().Unix()
	}
	if config.RestartStrategy == nil {
		config.RestartStrategy = &NoRestartStrategy{}
	}
	if config.NodesCount != len(config.Nodes) {
		panic("NodesCount != len(Nodes)")
	}
	if config.PartSize == 0 {
		panic("config.PartSize == 0")
	}
	if config.Timeout == 0 {
		panic("timeout should be set")
	}

	anyInfo := false
	for _, node := range config.Nodes {
		if node.HasFileInfo {
			anyInfo = true
		}
	}
	if !anyInfo {
		panic("No nodes with file info")
	}

	if config.SaveLogsPath == "" && testutil.LogDir != "" {
		now := time.Now()
		config.SaveLogsPath = filepath.Join(testutil.LogDir, now.Format("2006_01_02|15_04_05.000000000"))
	}
	if config.SaveLogsPath != "" {
		if err := os.MkdirAll(config.SaveLogsPath, 0777); err != nil {
			log.Panic("failed to create dir for logs", err)
		}
	}

	config.Gen = testutil.NewGenerator(config.Seed)
	config.IPs = config.Gen.GenerateIPs(config.NodesCount, config.Subnet)

	for i := 0; i < config.NodesCount; i++ {
		id := cli.DockerCreateContainer(config.Network, config.IPs[i])
		defer cli.DockerRmContainer(id)

		config.Nodes[i].IP = config.IPs[i]
		config.Nodes[i].DockerID = id

		if config.SaveLogsPath != "" {
			logFile := fmt.Sprintf("%02d_%s.log", i, id[:6])
			defer cli.DockerSaveLogs(config.Nodes[i].DockerID, filepath.Join(config.SaveLogsPath, logFile))
		}
	}

	localFile, err := os.CreateTemp("", "data.bin")
	if err != nil {
		log.Panic("failed to create temp file", err)
	}
	defer os.Remove(localFile.Name())
	defer localFile.Close()

	hashes, err := config.Gen.Generate(localFile, config.DataSize, config.PartSize)
	if err != nil {
		log.Panic("failed to generate data", err)
	}

	err = localFile.Sync()
	if err != nil {
		log.Panic("failed to sync data file")
	}

	config.Data = localFile
	config.Hashes = hashes
	config.Sha256Sum = sha256sum(config.Data)

	// will panic if something is wrong :)
	InitContainers(config)

	return RunTesting(config)
}

func InitContainers(config *Tester) {
	for i := 0; i < config.NodesCount; i++ {
		node := &config.Nodes[i]
		node.Config = GenerateConfig(config, node)
		configBytes := node.Config.Encode()
		cli.DockerCpStream(bytes.NewReader(configBytes), node.DockerID, "/torrent.conf")

		node.PartsMap = make(map[int]bool)
	}

	if config.RandomParts {
		for i := 0; i < len(config.Hashes); i++ {
			nodeID := config.Gen.Intn(config.NodesCount)
			config.Nodes[nodeID].PartsMap[i] = true
		}
	}

	for i := 0; i < config.NodesCount; i++ {
		InitData(config, &config.Nodes[i])
	}
}

func InitData(config *Tester, node *Node) {
	if node.HasFullData {
		err := cli.DockerCp(config.Data.Name(), node.DockerID+":/data.bin")
		if err != nil {
			log.Panic("failed to cp data to container", err)
		}
		return
	}

	if node.AllowExtraTrash {
		generateExtraTrash(config, node)
	}

	if config.RandomParts {
		generateFromPartsMap(config, node)
	}
}

func generateExtraTrash(config *Tester, node *Node) {
	var fileSize int

	switch val := config.Gen.Intn(4); val {
	case 0:
		fileSize = 0
	case 1:
		fileSize = config.DataSize
	case 2:
		fileSize = config.DataSize + 1
		if config.DataSize > 0 {
			fileSize += config.Gen.Intn(config.DataSize)
		}
	case 3:
		if fileSize > 0 {
			fileSize = config.Gen.Intn(config.DataSize)
		} else {
			fileSize = config.Gen.Intn(1024 * 1024)
		}
	}

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		_, err := config.Gen.Generate(w, fileSize, config.PartSize)
		if err != nil {
			log.Panic("failed to generate trash data", err)
		}
	}()

	cli.DockerCpStream(r, node.DockerID, "/data.bin")
}

type partsReader struct {
	tester   *Tester
	node     *Node
	nextPart int
	part     []byte
	tmp      []byte
}

func (r *partsReader) Read(b []byte) (n int, err error) {
	if len(r.part) != 0 {
		n = len(r.part)
		if len(b) < n {
			n = len(b)
		}
		copy(b, r.part[:n])
		r.part = r.part[n:]
		return n, nil
	}

	if r.nextPart >= len(r.tester.Hashes) {
		return 0, io.EOF
	}

	// fetch part from data file
	r.part = r.tmp
	n, _ = r.tester.Data.Read(r.part)

	// assert size
	if n < r.tester.PartSize && r.nextPart+1 < len(r.tester.Hashes) {
		panic("invalid read size")
	}

	if !r.node.PartsMap[r.nextPart] {
		// fill part with trash
		r.tester.Gen.Read(r.part)
	}
	r.nextPart++

	return r.Read(b)
}

func generateFromPartsMap(config *Tester, node *Node) {
	r := &partsReader{
		tester:   config,
		node:     node,
		nextPart: 0,
		part:     nil,
		tmp:      make([]byte, config.PartSize),
	}
	_, err := config.Data.Seek(0, 0)
	if err != nil {
		log.Panic("failed to seek file", err)
	}

	cli.DockerCpStream(r, node.DockerID, "/data.bin")
}

func GenerateConfig(config *Tester, node *Node) *testutil.Config {
	cfg := &testutil.Config{
		Peers:    nil,
		FileInfo: nil,
	}
	if node.HasAllPeers {
		peers := append([]string{}, config.IPs...)
		config.Gen.Shuffle(peers)
		cfg.Peers = peers
	}
	if node.HasFileInfo {
		cfg.FileInfo = &testutil.FileInfo{
			Size:     config.DataSize,
			PartSize: config.PartSize,
			Parts:    config.Hashes,
		}
	}
	if node.HasSinglePeer {
		cfg.Peers = append(cfg.Peers, config.Nodes[node.SinglePeerIndex].IP)
	}

	return cfg
}

func RunTesting(config *Tester) bool {
	var dockerIDs []string
	for _, node := range config.Nodes {
		dockerIDs = append(dockerIDs, node.DockerID)
	}

	config.RestartStrategy.Init(config, dockerIDs)
	log.Printf("Test is started, timeout=%v", config.Timeout.String())

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	go config.RestartStrategy.Run(ctx)

	shaOk := make(chan struct{})
	go checkForSha(ctx, shaOk, config)

await:
	select {
	case <-ctx.Done():
		log.Printf("WARN Timeout waiting for files to sync")
		break await
	case <-shaOk:
		log.Printf("INFO files look identical, stopping the test")
		break await
	}

	cancel()

	killChannel := make(chan struct{})
	go func() {
		defer close(killChannel)
		err := cli.DockerKill(dockerIDs...)
		if err != nil {
			log.Printf("WARN docker kill failed with error %v", err)
		}
	}()

	select {
	case <-time.After(time.Second * 60):
		log.Printf("FAIL Timeout waiting for containers to stop, test is considered failed")
		return false
	case <-killChannel:
		log.Printf("INFO containers are killed")
	}

	result := CheckFilesFull(config)
	if result {
		log.Printf("PASS files are identical")
	} else {
		log.Printf("FAIL files are not identical")
	}

	return result
}

func checkForSha(ctx context.Context, notifyShaOk chan struct{}, config *Tester) {
	defer close(notifyShaOk)

	i := 0
	ok := false
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		i, ok = CheckFilesFast(ctx, config, i)
		if ok {
			break
		}

		time.Sleep(time.Second)
	}
}

func CheckFilesFast(ctx context.Context, config *Tester, start int) (int, bool) {
	for i, node := range config.Nodes {
		select {
		case <-ctx.Done():
			return i, false
		default:
		}

		if i < start {
			continue
		}

		digest, err := cli.DockerExec(node.DockerID, "sha256sum", "/data.bin")
		if err != nil {
			log.Printf("WARN failed to get digest: %v", err)
			return i, false
		}

		tmp := strings.Split(digest, " ")
		if tmp[0] != config.Sha256Sum {
			return i, false
		}
	}

	return -1, true
}

func CheckFilesFull(config *Tester) bool {
	localFile, err := os.CreateTemp("", "tmp")
	if err != nil {
		log.Panic("failed to create temp file", err)
	}

	fileName := localFile.Name()
	defer os.Remove(fileName)
	localFile.Close()

	for _, node := range config.Nodes {
		err = cli.DockerCp(node.DockerID+":/data.bin", fileName)
		if err != nil {
			log.Panic("failed to copy data from container", err)
		}

		if !testutil.Compare(config.Data.Name(), fileName) {
			return false
		}
	}

	return true
}

func sha256sum(file *os.File) string {
	_, err := file.Seek(0, 0)
	if err != nil {
		log.Panic("failed to seek", err)
	}

	h := sha256.New()
	_, err = io.Copy(h, file)
	if err != nil {
		log.Panic("failed to copy sha256", err)
	}

	return hex.EncodeToString(h.Sum(nil))
}
