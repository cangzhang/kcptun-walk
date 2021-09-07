package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/lxn/walk"
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

func start(te *walk.TextEdit) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	binPath := ""
	if binPath, err = getBinPath(dir); err != nil {
		log.Println("get binary file error: ", err)

		binPath, err = download(dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("[kcptun] binary path is: ", binPath)
	runCmd(binPath, te)
}

func runCmd(bin string, te *walk.TextEdit) {
	var wg sync.WaitGroup
	args := []string{"-c", "/Users/al/tmp/kcptun/la.json"}
	cmd := exec.Command(bin, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
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
				te.AppendText(text + "\n")
			}
		}
	}()

	go func() {
		defer wg.Done()
		for errScanner.Scan() {
			text := errScanner.Text()
			log.Println(text)
			if len(text) > 0 {
				te.AppendText(text + "\n")
			}
		}
	}()

	wg.Wait()

	defer killCmd(cmd)
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func killCmd(cmd *exec.Cmd) {
	if err := cmd.Process.Kill(); err != nil {
		log.Fatal("failed to kill: ", cmd.Process.Pid)
	}
}

func download(dir string) (string, error) {
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

	workDir := filepath.Join(dir, "bin")
	_ = os.MkdirAll(workDir, os.ModePerm)
	p := filepath.Join(dir, "bin", obj.Name)
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
	p, err = ExtractTarGz(file, workDir)
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

func ExtractTarGz(gzipStream io.Reader, parentFolder string) (string, error) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed")
	}

	bin := ""
	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		p := filepath.Join(parentFolder, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(p, 0755); err != nil {
				log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(p)
			if err != nil {
				log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
			if strings.Contains(p, "client_") {
				bin = p
			}
			outFile.Close()

		default:
			log.Fatalf("ExtractTarGz: uknown type: %s in %s", header.Typeflag, header.Name)
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
