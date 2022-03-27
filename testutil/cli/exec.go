package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Exec(name string, arg ...string) *exec.Cmd {
	fmt.Printf("Running command : %s %s\n", name, arg)
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd
}

func ExecOutput(name string, arg ...string) (string, error) {
	fmt.Printf("Running command : %s %s\n", name, arg)
	cmd := exec.Command(name, arg...)

	b, err := cmd.Output()
	return strings.TrimSpace(string(b)), err
}
