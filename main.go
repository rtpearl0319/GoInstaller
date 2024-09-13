package main

import (
	"encoding/json"
	"fmt"
	"github.com/cespare/xxhash"
	"io"
	"net/http"
)

const (
	owner     = "rtpearl0319"
	repo      = "GoInstaller"
	dllName   = "GPSrvtTab.dll"
	addinName = "GPSrvtTab.addin"
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func main() {
	// GitHub API URL to get the latest release
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	// Fetch the latest release data
	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Printf("Error fetching latest release: %v\n", err)
		return
	}
	defer resp.Body.Close()

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

	hashAddin := xxhash.Sum64(addinData)
	hashDll := xxhash.Sum64(dllData)
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
		return nil, err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading file: %s", resp.Status)
	}

	// Get bytes
	byteData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return byteData, nil
}
