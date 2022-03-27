package prepenv

import (
	"os"
	"testutil"
	"time"
)

func GenerateFullFiles(dataFile, confFile string, dataSize, partSize int, peers []string) error {
	data, err := os.OpenFile(dataFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer data.Close()

	gen := testutil.NewGenerator(time.Now().Unix())
	hashes, err := gen.Generate(data, dataSize, partSize)
	if err != nil {
		return err
	}

	config := testutil.Config{
		Peers: peers,
		FileInfo: &testutil.FileInfo{
			Size:     dataSize,
			PartSize: partSize,
			Parts:    hashes,
		},
	}
	configBytes := config.Encode()

	err = os.WriteFile(confFile, configBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
