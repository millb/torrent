package cli

import (
	"io"
	"log"
	"os"
	"testutil"
)

// basic docker commands shouldn't fail. If they do, we should fail the test.

func DockerNetworkCreate(subnet string, name string) string {
	id, err := ExecOutput("docker", "network", "create", "--internal", "--subnet", subnet, name)
	if err != nil {
		log.Panic("failed to create network", err)
	}

	return string(id)
}

func DockerNetworkRm(id string) error {
	return Exec("docker", "network", "rm", id).Run()
}

func DockerCreateContainer(net string, ip string) string {
	id, err := ExecOutput("docker", "create", "--net", net, "--ip", ip, testutil.Image)
	if err != nil {
		log.Panic("failed to create container", err)
	}

	return id
}

func DockerRmContainer(id string) {
	err := Exec("docker", "rm", "-f", id).Run()
	if err != nil {
		log.Panic("failed to rm container", err)
	}
}

func DockerCpStream(r io.Reader, id string, dst string) {
	f, err := os.CreateTemp("", "cp")
	if err != nil {
		log.Panic("failed to create temp file", err)
	}
	defer os.Remove(f.Name())

	_, err = io.Copy(f, r)
	if err != nil {
		f.Close()
		log.Panic("failed to copy", err)
	}
	f.Close()

	err = DockerCp(f.Name(), id+":"+dst)
	if err != nil {
		log.Panic("failed to docker cp", err)
	}
}

func DockerCp(src string, dst string) error {
	return Exec("docker", "cp", src, dst).Run()
}

func DockerStart(id ...string) {
	err := Exec("docker", append([]string{"start"}, id...)...).Run()
	if err != nil {
		log.Panic("failed to docker start", err)
	}
}

func DockerRestart(id ...string) error {
	return Exec("docker", append([]string{"restart"}, id...)...).Run()
}

// DockerKill returns errors if the container is not running, I guess.
func DockerKill(id ...string) error {
	err := Exec("docker", append([]string{"kill"}, id...)...).Run()
	return err
}

func DockerExec(id string, cmd ...string) (string, error) {
	return ExecOutput("docker", append([]string{"exec", id}, cmd...)...)
}

func DockerSaveLogs(id string, dst string) {
	f, err := os.Create(dst)
	if err != nil {
		log.Printf("ERR failed to create file %s for logs, err=%v", dst, err)
		return
	}
	defer f.Close()

	cmd := Exec("docker", "logs", id, "--tail", "all")
	cmd.Stdout = f
	cmd.Stderr = f

	err = cmd.Run()
	if err != nil {
		log.Printf("ERR failed to save logs to file %s, err=%v", dst, err)
		return
	}
}
