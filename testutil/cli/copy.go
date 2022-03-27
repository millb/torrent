package cli

func CopyFile(src, dst string) error {
	return Exec("cp", src, dst).Run()
}
