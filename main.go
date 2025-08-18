package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

type PushEvent struct {
	Ref  string `json:"ref"`
	Repo struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

type RepoConfig struct {
	FullName  string
	Branch    string
	Secret    string
	ShellPath string
}

var repoConfigs []RepoConfig

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file: " + err.Error())
	}

	repoConfigs = loadConfigsFromEnv()
	if len(repoConfigs) == 0 {
		log.Fatal("No repository configurations found. Define REPO_FULL_NAME_1, GITHUB_WEBHOOK_SECRET_1, SHELL_PATH_1, etc.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5099"
	}

	// set up HTTP server
	http.HandleFunc("/wh", webhookHandler)

	// start the server on a specific port
	fmt.Println("Listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func loadConfigsFromEnv() []RepoConfig {
	var configs []RepoConfig
	for i := 1; ; i++ {
		fullName := os.Getenv(fmt.Sprintf("REPO_FULL_NAME_%d", i))
		if fullName == "" {
			break
		}
		secret := os.Getenv(fmt.Sprintf("GITHUB_WEBHOOK_SECRET_%d", i))
		shellPath := os.Getenv(fmt.Sprintf("SHELL_PATH_%d", i))
		branch := os.Getenv(fmt.Sprintf("BRANCH_NAME_%d", i))

		if secret == "" || shellPath == "" {
			log.Printf("Skipping repo config %d due to missing secret or shell path", i)
			continue
		}

		configs = append(configs, RepoConfig{
			FullName:  fullName,
			Branch:    branch,
			Secret:    secret,
			ShellPath: shellPath,
		})
	}
	return configs
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// read the body of the request (webhook payload)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// parse the request body to determine the repository
	pushEvent := &PushEvent{}
	err = json.Unmarshal(payload, pushEvent)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid payload: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// find matching repo configuration
	var cfg *RepoConfig
	for i := range repoConfigs {
		if repoConfigs[i].FullName == pushEvent.Repo.FullName {
			cfg = &repoConfigs[i]
			break
		}
	}
	if cfg == nil {
		log.Printf("Ignored push to repository: %s", pushEvent.Repo.FullName)
		fmt.Fprintf(w, "Ignored push to repository: %s", pushEvent.Repo.FullName)
		return
	}

	// verify the request signature using the repo-specific secret
	sigHeader := r.Header.Get("X-Hub-Signature-256")
	signature := strings.Split(sigHeader, "=")
	if len(signature) != 2 {
		http.Error(w, "Invalid signature, signature is too short", http.StatusUnauthorized)
		return
	}
	if !verifySignature(payload, cfg.Secret, signature[1]) {
		http.Error(w, "Invalid signature, signature mismatch", http.StatusUnauthorized)
		return
	}

	// optional branch check if provided
	if cfg.Branch != "" && pushEvent.Ref != "" {
		// GitHub push ref is typically "refs/heads/<branch>"
		if !strings.HasSuffix(pushEvent.Ref, "/"+cfg.Branch) {
			log.Printf("Ignored push to branch: %s (expecting %s)", pushEvent.Ref, cfg.Branch)
			fmt.Fprintf(w, "Ignored push to branch: %s", pushEvent.Ref)
			return
		}
	}

	// Create buffers to capture both stdout and stderr
	var outBuf, errBuf bytes.Buffer

	// execute the shell script
	cmd := exec.Command("bash", cfg.ShellPath)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing shell script: %s, stdout: %s, stderr: %s", err.Error(), outBuf.String(), errBuf.String()), http.StatusInternalServerError)
		return
	}

	// log the output of the shell script
	log.Printf("Shell script output: %s", outBuf.String())

	// return 200 OK status
	w.WriteHeader(http.StatusOK)
}

func verifySignature(payload []byte, githubWebhookSecret, signature string) bool {
	// compute HMAC-SHA256 hash of the payload using the secret key
	mac := hmac.New(sha256.New, []byte(githubWebhookSecret))
	mac.Write(payload)
	computedSignature := hex.EncodeToString(mac.Sum(nil))
	// compare the computed hash with the signature sent by GitHub
	return hmac.Equal([]byte(computedSignature), []byte(signature))
}
