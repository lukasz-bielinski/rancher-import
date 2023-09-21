package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"rancher-registration/rancherimport"
	"rancher-registration/vaultlogic"
	"time"
)

func logJSON(message string) {
	logMessage := map[string]string{"message": message}
	messageJSON, err := json.Marshal(logMessage)
	if err != nil {
		log.Fatalf("Could not marshal log message: %v", err)
	}
	log.Println(string(messageJSON))
}

func main() {
	RANCHER_SERVER := os.Getenv("RANCHER_SERVER")
	USERNAME := os.Getenv("USERNAME")
	if USERNAME == "" {
		USERNAME = "admin"
	}

	//TOKEN_TTL := os.Getenv("TOKEN_TTL")
	//if TOKEN_TTL == "" {
	//	TOKEN_TTL = "43200000"
	//}
	//
	//TOKEN_TTL_INT, err := strconv.Atoi(TOKEN_TTL)
	//if err != nil {
	//	// Handle the error
	//	log.Fatal("Invalid TOKEN_TTL value")
	//}

	//currentTime := time.Now().UTC().Format(time.RFC3339)
	//API_KEY_DESCRIPTION := fmt.Sprintf("Token created at %s with cronjob rancher-token", currentTime)

	var client *http.Client
	if os.Getenv("SKIP_TLS_VERIFY") == "true" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	for {
		resp, err := client.Get(fmt.Sprintf("https://%s", RANCHER_SERVER))
		if err == nil && resp.StatusCode == http.StatusOK {
			logJSON("Endpoint is accessible.")
			break
		} else {
			logJSON("Waiting for endpoint to be accessible...")
			time.Sleep(5 * time.Second)
		}
	}

	var API_KEY_NAME interface{}
	var API_KEY_TOKEN interface{}
	data := map[string]interface{}{
		"rancher2_access_key": API_KEY_NAME,
		"rancher2_secret_key": API_KEY_TOKEN,
	}

	secretData, err := vaultlogic.GetSecretWithKubernetesAuth(data)
	if err != nil {
		logJSON("Errot with getting API_KEY_NAME and API_KEY_TOKEN from Vault")
	}
	fmt.Println(secretData.AccessKey, secretData.SecretKey)

	err = rancherimport.ImportClusterToRancher(RANCHER_SERVER, secretData.AccessKey, secretData.SecretKey, client)
	if err != nil {
		log.Fatalf("Failed to import cluster to Rancher: %v", err)
	}

}
