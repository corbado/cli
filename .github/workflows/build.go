package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

func main() {
	tag := os.Getenv("tag")

	if err := os.Mkdir("dist", 0777); err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
	}

	targets := []struct {
		goos          string
		goarch        string
		fileExtension string
	}{
		{
			goos:          "linux",
			goarch:        "amd64",
			fileExtension: "",
		},
		{
			goos:          "linux",
			goarch:        "arm64",
			fileExtension: "",
		},
		{
			goos:          "darwin",
			goarch:        "amd64",
			fileExtension: "",
		},
		{
			goos:          "darwin",
			goarch:        "arm64",
			fileExtension: "",
		},
		{
			goos:          "windows",
			goarch:        "amd64",
			fileExtension: ".exe",
		},
		{
			goos:          "windows",
			goarch:        "arm64",
			fileExtension: ".exe",
		},
	}

	for _, target := range targets {
		if err := build(target.goos, target.goarch, target.fileExtension); err != nil {
			log.Fatalf("%+v", err)
		}

		if err := compress(tag, target.goos, target.goarch, target.fileExtension); err != nil {
			log.Fatalf("%+v", err)
		}
	}
}

func build(goos string, goarch string, fileExtension string) error {
	cmd := exec.Command("go", "build", "-buildmode", "exe", "-o", fmt.Sprintf("corbado%s", fileExtension), "./cmd/corbado")
	cmd.Env = []string{
		// Not inherited from calling environment, no idea why
		fmt.Sprintf("%s=%s", "GOMODCACHE", "/tmp"),
		// Not inherited from calling environment, no idea why
		fmt.Sprintf("%s=%s", "GOCACHE", "/tmp"),
		fmt.Sprintf("%s=%s", "GOOS", goos),
		fmt.Sprintf("%s=%s", "GOARCH", goarch),
		fmt.Sprintf("%s=%s", "CGO_ENABLED", "0"),
	}

	var stdoutBuffer, stderrBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer

	if err := cmd.Run(); err != nil {
		return errors.WithMessagef(err, "Stdout: %s, Stderr: %s", stdoutBuffer.String(), stderrBuffer.String())
	}

	return nil
}

func compress(tag string, goos string, goarch string, fileExtension string) error {
	output := "tar.gz"

	if goos == "darwin" {
		goos = "macos"
	} else if goos == "windows" {
		output = "zip"
	}

	var cmd *exec.Cmd
	if output == "tar.gz" {
		cmd = exec.Command("tar", "czf", fmt.Sprintf("dist/corbado_cli_%s_%s_%s.tar.gz", tag, goos, goarch), fmt.Sprintf("corbado%s", fileExtension))
	} else {
		cmd = exec.Command("zip", fmt.Sprintf("dist/corbado_cli_%s_%s_%s.zip", tag, goos, goarch), fmt.Sprintf("corbado%s", fileExtension))
	}

	var stdoutBuffer, stderrBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer

	if err := cmd.Run(); err != nil {
		return errors.WithMessagef(err, "Stdout: %s, Stderr: %s", stdoutBuffer.String(), stderrBuffer.String())
	}

	return nil
}
