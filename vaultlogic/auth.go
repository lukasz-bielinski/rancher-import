package vaultlogic

import (
	"context"
	"encoding/json"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"
	"log"
	"os"
)

// LogJSON logs a JSON-formatted message
func logJSON(message string) {
	logMessage := map[string]string{"message": message}
	messageJSON, err := json.Marshal(logMessage)
	if err != nil {
		log.Fatalf("Could not marshal log message: %v", err)
	}
	log.Println(string(messageJSON))
}

// LogAndReturnError logs an error message and returns an error
func logAndReturnError(errMessage string, originalErr error) error {
	if originalErr != nil {
		errMessage = fmt.Sprintf("%s: %s", errMessage, originalErr.Error())
	}
	logJSON(errMessage)
	return fmt.Errorf(errMessage)
}

// SecretData represents the data fetched from Vault
type SecretData struct {
	AccessKey string `json:"rancher2_access_key"`
	SecretKey string `json:"rancher2_secret_key"`
}

// GetSecretWithKubernetesAuth gets a secret from Vault using Kubernetes authentication
func GetSecretWithKubernetesAuth(keys map[string]interface{}) (*SecretData, error) {
	VAULT_ADDR := os.Getenv("VAULT_ADDR")
	if len(VAULT_ADDR) == 0 {
		return nil, logAndReturnError("VAULT_ADDR not set", nil)
	}

	secretEngine := os.Getenv("VAULT_SECRET_ENGINE")
	if secretEngine == "" {
		secretEngine = "kv-v2"
	}

	secretPath := os.Getenv("VAULT_SECRET_PATH")
	if secretPath == "" {
		secretPath = "creds"
	}

	config := &vault.Config{
		Address: "http://" + VAULT_ADDR + ":8200",
	}

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, logAndReturnError("Unable to initialize Vault client", err)
	}

	k8sAuth, err := auth.NewKubernetesAuth(
		"rancher-tokens",
		auth.WithServiceAccountTokenPath("/var/run/secrets/kubernetes.io/serviceaccount/token"),
	)
	if err != nil {
		return nil, logAndReturnError("Unable to initialize Kubernetes auth method", err)
	}

	authInfo, err := client.Auth().Login(context.TODO(), k8sAuth)
	if err != nil {
		return nil, logAndReturnError("Unable to log in with Kubernetes auth", err)
	}
	if authInfo == nil {
		return nil, logAndReturnError("No auth info was returned after login", nil)
	}

	kv := client.KVv2(secretEngine)
	secret, err := kv.Get(context.Background(), secretPath)

	if err != nil {
		return nil, logAndReturnError("Unable to read secret", err)
	}

	accessKey, ok1 := secret.Data["rancher2_access_key"].(string)
	secretKey, ok2 := secret.Data["rancher2_secret_key"].(string)

	if !ok1 || !ok2 {
		return nil, logAndReturnError("Value type assertion failed", nil)
	}

	logJSON("Successfully fetched secret")
	return &SecretData{
		AccessKey: accessKey,
		SecretKey: secretKey,
	}, nil
}
