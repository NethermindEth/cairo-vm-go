//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunBenchmarks() {
	benchmarkDirs := []string{
		"./pkg/hintrunner",
		"./pkg/vm",
		"./pkg/runners/zero",
	}

	// create benchmark result folder
	branchName, commitHash, date := getGitInfo()
	dirName := fmt.Sprintf("%s %s %s", branchName, date, commitHash)
	dirPath := filepath.Join("benchmarks", dirName)

	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Fatalf("creating parent benchmark folder: %s", err)
	}

	for _, pkgDir := range benchmarkDirs {
		pkg := filepath.Base(pkgDir)

		benchPath := filepath.Join(dirPath, pkg)
		err := os.MkdirAll(benchPath, 0755)
		if err != nil {
			log.Fatalf("creating benchmark subfolder for pkg %s: %s", pkg, err)
		}

		cpuPath := filepath.Join(benchPath, "cpu.out")
		memPath := filepath.Join(benchPath, "mem.out")

		cmd := exec.Command(
			"go", "test", pkgDir, "-bench", ".",
			"-cpuprofile", cpuPath, "-memprofile", memPath, "-benchmem",
		)
		var stdOut bytes.Buffer
		cmd.Stdout = &stdOut
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			log.Fatalf("command failed: %s", err)
		}

		stdOutPath := filepath.Join(benchPath, "stdout.txt")
		err = os.WriteFile(stdOutPath, stdOut.Bytes(), 0644)

		fmt.Printf("%s benchmark complete\n", pkg)
	}
	fmt.Printf("\nBenchmarking complete. Results stored in:\n\"%s\"\n", dirPath)
}

func getGitInfo() (string, string, string) {
	branchBytes, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		log.Fatalf("branch name parsing: %s\n%s", err, string(branchBytes))
	}

	commitHashBytes, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		log.Fatalf("commit hash parsing: %s\n%s", err, string(commitHashBytes))
	}

	dateBytes, err := exec.Command("git", "log", "-1", "--format=%cd", "--date=short").CombinedOutput()
	if err != nil {
		log.Fatalf("date parsing: %s\n%s", err, string(dateBytes))
	}

	return strings.TrimSpace(string(branchBytes)),
		strings.TrimSpace(string(commitHashBytes))[0:8],
		strings.TrimSpace(string(dateBytes))
}

func main() {
	RunBenchmarks()
}
