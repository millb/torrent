package tasks

import (
	"log"
	"os"
	"testutil"
	"testutil/cli"
	"testutil/prepenv"
	"time"
)

func SimpleDockerComposeTest() bool {
	cli.ComposeBuild()
	cli.ComposeDown()

	// clean up on exit
	defer cli.ComposeDown()

	const (
		conf1    = "./tmp/torrent1.conf"
		conf2    = "./tmp/torrent2.conf"
		data1    = "./tmp/data1.bin"
		data2    = "./tmp/data2.bin"
		dataSize = 2048
		partSize = 256
	)

	peers := []string{"172.31.111.42", "172.31.111.69"}
	if err := prepenv.GenerateFullFiles(data1, conf1, dataSize, partSize, peers); err != nil {
		log.Printf("FAIL failed to generate files for test")
		return false
	}

	if err := os.WriteFile(data2, []byte(""), 0644); err != nil {
		log.Printf("FAIL failed to generate files for test")
		return false
	}

	if err := cli.CopyFile(conf1, conf2); err != nil {
		log.Printf("FAIL failed to copy config file")
		return false
	}

	// start the test
	cli.ComposeUp()

	timeout := time.After(30 * time.Second)

await:
	for {
		select {
		case <-timeout:
			log.Printf("WARN Timeout waiting for files to sync")
			break await
		default:
			if testutil.Compare(data1, data2) {
				log.Printf("INFO files are identical, stopping the test")
				break await
			}
		}
	}

	result := testutil.Compare(data1, data2)
	if result {
		log.Printf("PASS files are identical")
	} else {
		log.Printf("FAIL files are not identical")
	}
	return result
}
