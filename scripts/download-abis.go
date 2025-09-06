package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// ContractDeployment specifies the deployment fields we want to extract.
type ContractDeployment struct {
	ABI json.RawMessage `json:"abi"`
}

// downloadABI downloads the ABI for a given contract name in specified output path.
func downloadABI(contractName, outputPath string) error {
	url := fmt.Sprintf("https://raw.githubusercontent.com/livepeer/protocol/delta/deployments/arbitrumMainnet/%s.json", contractName)

	fmt.Printf("Downloading %s ABI from %s...\n", contractName, url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %v", contractName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download %s: HTTP %d", contractName, resp.StatusCode)
	}

	var deployment ContractDeployment
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response for %s: %v", contractName, err)
	}

	if err := json.Unmarshal(body, &deployment); err != nil {
		return fmt.Errorf("failed to parse JSON for %s: %v", contractName, err)
	}

	// Create directory if it doesn't exist.
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Write just the ABI part to file.
	if err := os.WriteFile(outputPath, deployment.ABI, 0644); err != nil {
		return fmt.Errorf("failed to write ABI file %s: %v", outputPath, err)
	}

	fmt.Printf("✅ Downloaded %s ABI to %s\n", contractName, outputPath)
	return nil
}

func main() {
	contracts := map[string]string{
		"BondingManagerTarget": "../ABIs/BondingManager.json",
		"RoundsManagerTarget":  "../ABIs/RoundsManager.json",
	}

	fmt.Println("Downloading Livepeer protocol ABIs...")

	for contractName, outputPath := range contracts {
		if err := downloadABI(contractName, outputPath); err != nil {
			fmt.Printf("❌ Error downloading %s: %v\n", contractName, err)
			os.Exit(1)
		}
	}

	fmt.Println("✅ All ABIs downloaded successfully!")
}
