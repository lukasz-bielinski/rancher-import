package rancherimport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type ClusterRegistrationTokenStatus struct {
	InsecureCommand    string `json:"insecureCommand"`
	Command            string `json:"command"`
	WindowsNodeCommand string `json:"windowsNodeCommand"`
	NodeCommand        string `json:"nodeCommand"`
	ManifestURL        string `json:"manifestUrl"`
	Token              string `json:"token"`
}

func ImportClusterToRancher(rancherServer, apiKeyName, apiKeyToken string, httpClient *http.Client) error {
	type ClusterCreateRequest struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	var body []byte
	payload := &ClusterCreateRequest{
		Name: fmt.Sprintf("cluster-1"),
		Type: "import",
	}

	jsonData, _ := json.Marshal(payload)

	exists, err := doesClusterExist(payload.Name, rancherServer, apiKeyToken, httpClient)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("cluster with name cluster-1 already exists in Rancher")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/v3/clusters", rancherServer), bytes.NewBuffer(jsonData))
	time.Sleep(3 * time.Second)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKeyToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ = ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to create cluster in Rancher: %s", body)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("Rancher response: %+v\n", result)

	// Fetch the clusterRegistrationTokens link from Rancher's response
	tokensURL, ok := result["links"].(map[string]interface{})["clusterRegistrationTokens"].(string)
	if !ok {
		return fmt.Errorf("no clusterRegistrationTokens link found in Rancher's response")
	}

	// Fetch the tokens data from Rancher
	time.Sleep(3 * time.Second)

	req, _ = http.NewRequest("GET", tokensURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKeyToken))
	resp, err = httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch cluster registration tokens data from Rancher: %v", err)
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	var tokensResponse map[string]interface{}
	json.Unmarshal(body, &tokensResponse)

	fmt.Printf("Data type: %T, content: %+v\n", tokensResponse["data"], tokensResponse["data"])
	tokensData, exists := tokensResponse["data"].([]interface{})
	if !exists || len(tokensData) == 0 {
		return fmt.Errorf("no tokens data found in Rancher's cluster registration tokens response")
	}

	firstToken := tokensData[0].(map[string]interface{})
	manifestUrl, urlExists := firstToken["manifestUrl"].(string)
	if !urlExists || manifestUrl == "" {
		return fmt.Errorf("manifestUrl not found in the cluster registration token data")
	}

	// Fetch the manifest content
	req, _ = http.NewRequest("GET", manifestUrl, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKeyToken))
	resp, err = httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest content from URL: %s, error: %v", manifestUrl, err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read the content from the fetched URL: %s, error: %v", manifestUrl, err)
	}
	manifestContent := string(body)
	if manifestContent == "" {
		return fmt.Errorf("manifest content is empty from URL: %s", manifestUrl)
	}

	// Use kubectl to apply the manifest content
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(manifestContent)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to apply Rancher import manifest: %v, stderr: %s", err, stderr.String())
	}

	return nil
}
func doesClusterExist(clusterName, rancherServer, apiKeyToken string, httpClient *http.Client) (bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/v3/clusters?name=%s", rancherServer, clusterName), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKeyToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return false, fmt.Errorf("failed to get clusters from Rancher: %s", body)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if dataInterface, ok := result["data"]; ok {
		data, ok := dataInterface.([]interface{})
		if ok && len(data) > 0 {
			return true, nil
		}
	}

	return false, nil
}
