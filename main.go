package main

import (
	"encoding/json"
	"fmt"
	"github.com/cespare/xxhash"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

const (
	owner     = "rtpearl0319"
	repo      = "GPSrvtTab"
	dllName   = "GPSrvtTabWrapper.dll"
	addinName = "GPSTab.addin"
)

var (
	versions = []string{"2025", "2024", "2023", "2022", "2021"}
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func main() {

	// Check if windows
	if os.Getenv("OS") != "Windows_NT" {
		fmt.Println("This script is only for Windows")
		return
	}

	// Get the ProgramData folder path from the environment variable
	programDataFolder := os.Getenv("ProgramData")

	// If the environment variable is empty, use a fallback path
	if programDataFolder == "" {
		programDataFolder = filepath.Join(os.Getenv("SystemDrive"), "ProgramData")
	}

	revitPath := path.Join(programDataFolder, "Autodesk", "Revit")

	// GitHub API URL to get the latest release
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	// Fetch the latest release data
	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Printf("Error fetching latest release: %v\n", err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: received non-OK HTTP status %d\n", resp.StatusCode)
		return
	}
	// Decode the JSON response
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return
	}
	// Download the Addin file
	addinData, err := downloadAddin(release)
	if err != nil {
		fmt.Printf("Error downloading Addin: %v\n", err)
		return
	}
	// Download the DLL file
	dllData, err := downloadDLL(release)
	if err != nil {
		fmt.Printf("Error downloading DLL: %v\n", err)
		return
	}

	if err = installAddin(path.Join(revitPath, `Addins`), addinData); err != nil {
		fmt.Printf("Error Installing Addin: %v\n", err)
		return
	}

	if err = installDLL(path.Join(revitPath, "DLLs"), dllData); err != nil {
		fmt.Printf("Error Installing DLL: %v\n", err)
		return
	}
}

func installAddin(addinsPath string, addinData []byte) error {

	hashAddinOnline := xxhash.Sum64(addinData)

	for _, version := range versions {

		if _, err := os.Stat(path.Join(addinsPath, version)); os.IsNotExist(err) {
			continue
		}

		addinPath := path.Join(addinsPath, version, addinName)
		if _, err := os.Stat(addinPath); !os.IsNotExist(err) {

			dat, err := os.ReadFile(addinPath)
			if err != nil {
				return fmt.Errorf("error reading Addin file: %v", err)
			}

			hashAddinLocal := xxhash.Sum64(dat)

			if hashAddinOnline == hashAddinLocal {
				continue
			}
		}
		if err := os.WriteFile(addinPath, addinData, 0777); err != nil {
			return fmt.Errorf("error writing Addin file: %v", err)
		}
	}
	return nil
}

func installDLL(dllsPath string, dllData []byte) error {

	hashDllOnline := xxhash.Sum64(dllData)

	if _, err := os.Stat(dllsPath); os.IsNotExist(err) {
		err := os.Mkdir(dllsPath, 0777)
		if err != nil {
			return fmt.Errorf("error creating DLL directory: %v", err)
		}
	}

	dllPath := path.Join(dllsPath, `GPSrvtTab.dll`)

	if _, err := os.Stat(dllPath); !os.IsNotExist(err) {
		dat, err := os.ReadFile(dllPath)
		if err != nil {
			return fmt.Errorf("error reading DLL file: %v", err)
		}

		hashDllLocal := xxhash.Sum64(dat)

		if hashDllOnline == hashDllLocal {
			return nil
		}
	}

	if err := os.WriteFile(dllPath, dllData, 0777); err != nil {
		return fmt.Errorf("error writing DLL file: %v", err)
	}

	return nil
}

func downloadAddin(release Release) ([]byte, error) {

	// Find the Addin asset
	var addinAsset string
	for _, asset := range release.Assets {
		if asset.Name == addinName {
			addinAsset = asset.BrowserDownloadURL
			break
		}
	}

	if addinAsset == "" {
		return nil, fmt.Errorf("addin file not found in latest release")
	}

	// Download the Addin file
	data, err := downloadFile(addinAsset)
	if err != nil {
		return nil, fmt.Errorf("error downloading Addin: %v", err)
	}

	return data, nil
}

func downloadDLL(release Release) ([]byte, error) {

	// Find the DLL asset
	var dllAsset string
	for _, asset := range release.Assets {
		if asset.Name == dllName { // Replace with actual DLL file name
			dllAsset = asset.BrowserDownloadURL
			break
		}
	}

	if dllAsset == "" {
		return nil, fmt.Errorf("DLL file not found in latest release")
	}

	// Download the DLL file
	data, err := downloadFile(dllAsset)
	if err != nil {
		return nil, fmt.Errorf("error downloading DLL: %v", err)
	}

	return data, nil
}

// Helper function to download a file
func downloadFile(url string) ([]byte, error) {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error downloading file: %v\n", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading file: %s", resp.Status)
	}

	// Get bytes
	byteData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return byteData, nil
}
