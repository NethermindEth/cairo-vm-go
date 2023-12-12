//go:build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunBenchmarks(pkgSubstr, testSubstr string) {
	pkgPrefix := filepath.Join(".", "pkg")
	benchmarkDirs := []string{
		filepath.Join(pkgPrefix, "hintrunner"),
		filepath.Join(pkgPrefix, "vm"),
		filepath.Join(pkgPrefix, "runners", "zero"),
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
		if !strings.Contains(pkg, pkgSubstr) {
			continue
		}

		pkgDir, err := filepath.Abs(pkgDir)
		if err != nil {
			log.Fatalf("locating pkg %s", pkg)
		}

		benchPath := filepath.Join(dirPath, pkg)
		err = os.MkdirAll(benchPath, 0755)
		if err != nil {
			log.Fatalf("creating benchmark subfolder for pkg %s: %s", pkg, err)
		}

		cpuPath := filepath.Join(benchPath, "cpu.out")
		memPath := filepath.Join(benchPath, "mem.out")

		cmd := exec.Command(
			"go", "test", pkgDir, "-bench", testSubstr,
			"-cpuprofile", cpuPath, "-memprofile", memPath, "-benchmem",
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("failed to benchmark %s: %s\n%s", pkg, err, string(output))
		}

		stdOutPath := filepath.Join(benchPath, "stdout.txt")
		err = os.WriteFile(stdOutPath, output, 0644)

		fmt.Printf(" - %s ✔\n", pkg)
	}
	fmt.Printf("Done! Results stored in: \"%s\"\n", dirPath)
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
	var pkgName string
	var testName string
	flag.StringVar(
		&pkgName, "pkg", "", "package names that contain this substring will be benchmarked",
	)
	flag.StringVar(
		&testName, "test", ".", "test names that contain this substring will be benchmarked",
	)
	flag.Parse()

	RunBenchmarks(pkgName, testName)
}
