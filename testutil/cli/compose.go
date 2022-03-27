package cli

import "log"

// docker-compose commands shouldn't fail. If they do, we should fail the test.

func ComposeDown() {
	err := Exec("docker-compose", "down", "-v").Run()
	if err != nil {
		log.Panic("failed to compose down", err)
	}
}

func ComposeBuild() {
	err := Exec("docker-compose", "build").Run()
	if err != nil {
		log.Panic("failed to build compose", err)
	}
}

func ComposeUp() {
	err := Exec("docker-compose", "up", "-d").Run()
	if err != nil {
		log.Panic("failed to up compose", err)
	}
}
