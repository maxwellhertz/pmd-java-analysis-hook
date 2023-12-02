package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/go-git/go-git/v5"
)

const (
	PmdReleasePkg     = "pmd-bin-6.55.0"
	PmdReleaseUrl     = "https://github.com/pmd/pmd/releases/download/pmd_releases%2F6.55.0/pmd-bin-6.55.0.zip"
	PmdLocalPath      = "./target/pmd"
	PmdLocalBinPath   = PmdLocalPath + "/" + PmdReleasePkg + "/" + "bin"
	PmdDefaultRuleSet = "rulesets/java/quickstart.xml"

	AppName = "pmd-java-pre-commit-hook"
)

func main() {
	stagedJavaFiles, err := getStagedJavaFiles()
	if err != nil {
		log.Printf("[WARN] %v failed to get staged java files: %v\n", AppName, err)
		os.Exit(0)
	}
	if len(stagedJavaFiles) == 0 {
		log.Printf("[INFO] %v didn't find any staged Java files\n", AppName)
		os.Exit(0)
	}

	pmdTargetFile := filepath.Join(PmdLocalPath, "sourceFiles.txt")
	file, err := createFile(pmdTargetFile)
	if err != nil {
		log.Printf("[WARN] %v failed to create %v: %v\n", AppName, pmdTargetFile, err)
		os.Exit(0)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, path := range stagedJavaFiles {
		writer.WriteString(path + "\n")
	}
	writer.Flush()

	ok, pmdBinPath := lookupPmd()
	if !ok {
		log.Printf("[INFO] %v is installing PMD...\n", AppName)
		pmdBinPath, err = installPmd()
		if err != nil {
			log.Printf("[WARN] %v failed to install PMD: %v\n", AppName, err)
			os.Exit(0)
		}
	}

	args := os.Args
	ruleSet := PmdDefaultRuleSet
	if len(args) > 1 {
		ruleSet = args[1]
	}
	pmdCmd := exec.Command(filepath.Join(pmdBinPath, getPmdScript()), getPmdCommand(), "-R", ruleSet, "-f", "text", "--cache", filepath.Join(PmdLocalPath, "cache"), "--file-list", pmdTargetFile)
	pmdCmd.Stdout = os.Stdout
	pmdCmd.Stderr = os.Stderr
	err = pmdCmd.Run()
	if err == nil {
		os.Exit(0)
	}
	os.Exit(1)
}

func getStagedJavaFiles() ([]string, error) {
	repo, err := git.PlainOpen(".")
	if errors.Is(err, git.ErrRepositoryNotExists) {
		return make([]string, 0), nil
	}
	workTree, err := repo.Worktree()
	if err != nil {
		return make([]string, 0), err
	}
	status, err := workTree.Status()
	if err != nil {
		return make([]string, 0), err
	}

	stagedJavaFiles := make([]string, 0, len(status))
	for srcFilePath, fileStatus := range status {
		if fileStatus.Staging != git.Unmodified && fileStatus.Staging != git.Untracked && filepath.Ext(srcFilePath) != ".java" {
			continue
		}
		stagedJavaFiles = append(stagedJavaFiles, srcFilePath)
	}
	if len(stagedJavaFiles) == 0 {
		return make([]string, 0), nil
	}
	return stagedJavaFiles, nil
}

func lookupPmd() (bool, string) {
	fileInfo, err := os.Stat(PmdLocalBinPath)
	if os.IsNotExist(err) {
		return false, ""
	}
	if fileInfo != nil {
		return true, PmdLocalBinPath
	}
	return false, ""
}

func installPmd() (string, error) {
	resp, err := http.Get(PmdReleaseUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	pmdZip := PmdLocalPath + "pkg"
	out, err := createFile(pmdZip)
	if err != nil {
		return "", err
	}
	defer func() {
		out.Close()
		os.Remove(pmdZip)
	}()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	unzip := exec.Command("unzip", pmdZip, "-d", PmdLocalPath)
	if err := unzip.Run(); err != nil {
		return "", err
	}
	return filepath.Join(PmdLocalPath, PmdReleasePkg, "bin"), nil
}

func createFile(path string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, err
	}
	return os.Create(path)
}

func getPmdScript() string {
	switch runtime.GOOS {
	case "windows":
		return "pmd.bat"
	default:
		return "run.sh"
	}
}

func getPmdCommand() string {
	switch runtime.GOOS {
	case "windows":
		return ""
	default:
		return "pmd"
	}
}
