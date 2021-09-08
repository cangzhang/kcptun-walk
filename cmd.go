// +build windows

package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
)

const (
	latestReleaseUrl = "https://api.github.com/repos/xtaci/kcptun/releases/latest"
	WinPkg           = "-windows-amd64-"
	LinuxPkg         = "-linux-amd64-"
	MacPkg           = "-darwin-amd64-"
)

func startCmd(config *Config) {
	dir, err := os.Getwd()
	if err != nil {
		config.logToTextarea(err.Error())
		return
	}

	config.pwd = dir
	config.binDir = filepath.Join(config.pwd, "bin")
	binPath := ""
	if binPath, err = getBinPath(dir); err != nil {
		log.Println("get local binary file failed, ", err)
		config.logToTextarea("try to download from github.com")

		binPath, err = download(config)
		if err != nil {
			config.logToTextarea(err.Error())
			return
		}
	}

	config.binPath = binPath
	log.Println("[kcptun] binary path is: ", binPath)
	config.jsonPath = filepath.Join(config.pwd, "config.json")
	runCmd(binPath, config)
}

func runCmd(bin string, config *Config) {
	var wg sync.WaitGroup
	args := []string{"-c", config.jsonPath}
	cmd := exec.Command(bin, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		config.logToTextarea(err.Error())
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		config.logToTextarea(err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		config.logToTextarea(err.Error())
		return
	}

	outScanner := bufio.NewScanner(stdout)
	errScanner := bufio.NewScanner(stderr)

	wg.Add(2)
	go func() {
		defer wg.Done()
		for outScanner.Scan() {
			text := outScanner.Text()
			log.Println(text)
			if len(text) > 0 {
				config.logToTextarea(text)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for errScanner.Scan() {
			text := errScanner.Text()
			log.Println(text)
			if len(text) > 0 {
				config.logToTextarea(text)
			}
		}
	}()

	wg.Wait()
	config.cmd = cmd

	defer killCmd(config)
	if err := cmd.Wait(); err != nil {
		config.cmd = nil
		config.logToTextarea(err.Error())
		return
	}
}

func download(config *Config) (string, error) {
	config.logToTextarea("[fetch] fetching latest binary...")
	resp, err := http.Get(latestReleaseUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	result := IReleaseResp{}
	_ = json.Unmarshal(body, &result)
	var obj IReleaseAsset
	for _, asset := range result.Assets {
		if strings.Contains(asset.Name, getTargetPkgName()) {
			obj = asset
			break
		}
	}

	config.logToTextarea("[fetch] got latest version " + result.TagName)

	r, err := http.Get(obj.BrowserDownloadURL)
	if err != nil {
		return "", err
	}
	if r.StatusCode != 200 {
		return "", errors.New(r.Status)
	}
	defer r.Body.Close()

	_ = os.MkdirAll(config.binDir, os.ModePerm)
	p := filepath.Join(config.binDir, obj.Name)
	out, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err = io.Copy(out, r.Body); err != nil {
		return "", err
	}
	config.logToTextarea("[fetch] downloaded: " + obj.Name)

	file, err := os.Open(p)
	if err != nil {
		return "", err
	}

	config.logToTextarea("[fetch] prepare to decompress " + obj.Name)

	p, err = extractTar(file, config)
	defer file.Close()
	if err != nil {
		return "", err
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(p, 0755); err != nil {
			return "", err
		}
	}

	config.logToTextarea("[fetch] decompressed")
	return p, nil
}

func extractTar(gzipStream io.Reader, config *Config) (string, error) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		config.logToTextarea("ExtractTarGz: NewReader failed")
		return "", err
	}

	bin := ""
	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			config.logToTextarea("ExtractTarGz: NewReader failed " + err.Error())
			return "", err
		}

		p := filepath.Join(config.binDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(p, 0755); err != nil {
				config.logToTextarea("ExtractTarGz: Mkdir() failed: " + err.Error())
				return "", err
			}
		case tar.TypeReg:
			outFile, err := os.Create(p)
			if err != nil {
				config.logToTextarea("ExtractTarGz: Create() failed: " + err.Error())
				return "", err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				config.logToTextarea("ExtractTarGz: Copy() failed: " + err.Error())
				return "", err
			}
			if strings.Contains(p, "client_") {
				bin = p
			}
			outFile.Close()

		default:
			config.logToTextarea("ExtractTarGz: unknown type:" + string(header.Typeflag) + " " + header.Name)
			return "", err
		}

	}

	return bin, nil
}

func getTargetPkgName() string {
	switch runtime.GOOS {
	case "windows":
		return WinPkg
	case "linux":
		return LinuxPkg
	case "darwin":
		return MacPkg
	default:
		log.Fatalf("No platform found.")
	}
	return ""
}

func getBinPath(dir string) (string, error) {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	filePath := filepath.Join(dir, "bin", "client_"+runtime.GOOS+"_amd64"+ext)
	file, err := os.Open(filePath)
	defer file.Close()

	if err == os.ErrNotExist {
		return "", err
	}
	if err != nil {
		return "", err
	}

	cmd := exec.Command(filePath, "-v")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	log.Printf(string(out))
	return filePath, nil
}
