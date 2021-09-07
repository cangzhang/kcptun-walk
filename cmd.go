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
)

const (
	latestReleaseUrl = "https://api.github.com/repos/xtaci/kcptun/releases/latest"
	WinPkg           = "-windows-amd64-"
	LinuxPkg         = "-linux-amd64-"
	MacPkg           = "-darwin-amd64-"
)

func start(config *Config) {
	dir, err := os.Getwd()
	if err != nil {
		log.Println(err)
		config.textEdit.AppendText(err.Error())
		return
	}

	config.pwd = dir
	binPath := ""
	if binPath, err = getBinPath(dir); err != nil {
		log.Println("get local binary file failed, ", err)

		binPath, err = download(config)
		if err != nil {
			log.Println(err)
			return
		}
	}

	config.binPath = binPath
	log.Println("[kcptun] binary path is: ", binPath)
	config.jsonPath = filepath.Join(dir, "config.json")
	runCmd(binPath, config)
}

func runCmd(bin string, config *Config) {
	var wg sync.WaitGroup
	args := []string{"-c", config.jsonPath}
	cmd := exec.Command(bin, args...)
	config.cmd = cmd
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		config.textEdit.AppendText(err.Error())
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		config.textEdit.AppendText(err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		config.textEdit.AppendText(err.Error())
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
				config.textEdit.AppendText(text + "\n")
			}
		}
	}()

	go func() {
		defer wg.Done()
		for errScanner.Scan() {
			text := errScanner.Text()
			log.Println(text)
			if len(text) > 0 {
				config.textEdit.AppendText(text + "\n")
			}
		}
	}()

	wg.Wait()

	defer killCmd(config)
	if err := cmd.Wait(); err != nil {
		config.textEdit.AppendText(err.Error())
		return
	}
}

func killCmd(config *Config) {
	err := config.cmd.Process.Kill()
	if err != nil {
		config.textEdit.AppendText("failed to kill: " + string(config.cmd.Process.Pid))
		return
	}

	config.textEdit.AppendText("killed")
	return
}

func download(config *Config) (string, error) {
	log.Println("fetching latest binary...")
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
	log.Printf("got latest version %s", result.TagName)
	r, err := http.Get(obj.BrowserDownloadURL)
	if err != nil {
		return "", err
	}
	if r.StatusCode != 200 {
		return "", errors.New(r.Status)
	}
	defer r.Body.Close()

	config.binDir = filepath.Join(config.pwd, "bin")
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
	log.Printf("downloaded tar %s", obj.Name)

	file, err := os.Open(p)
	if err != nil {
		return "", err
	}

	log.Printf("prepare to decompress %s", obj.Name)
	p, err = extractTar(file, config)
	if err != nil {
		return "", err
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(p, 0755); err != nil {
			return "", err
		}
	}

	log.Printf("decompressed")
	return p, nil
}

func extractTar(gzipStream io.Reader, config *Config) (string, error) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		config.textEdit.AppendText("ExtractTarGz: NewReader failed")
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
			config.textEdit.AppendText("ExtractTarGz: NewReader failed " + err.Error())
			return "", err
		}

		p := filepath.Join(config.binDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(p, 0755); err != nil {
				config.textEdit.AppendText("ExtractTarGz: Mkdir() failed: " + err.Error())
				return "", err
			}
		case tar.TypeReg:
			outFile, err := os.Create(p)
			if err != nil {
				config.textEdit.AppendText("ExtractTarGz: Create() failed: " + err.Error())
				return "", err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				config.textEdit.AppendText("ExtractTarGz: Copy() failed: " + err.Error())
				return "", err
			}
			if strings.Contains(p, "client_") {
				bin = p
			}
			outFile.Close()

		default:
			config.textEdit.AppendText("ExtractTarGz: unknown type:" + string(header.Typeflag) + " " + header.Name)
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
	filePath := filepath.Join(dir, "bin", "client_"+runtime.GOOS+"_amd64")
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	defer file.Close()

	if err == os.ErrNotExist {
		return "", err
	}
	if err != nil {
		return "", err
	}

	out, err := exec.Command(filePath, "-v").Output()
	if err != nil {
		return "", err
	}

	log.Printf(string(out))
	return filePath, nil
}
