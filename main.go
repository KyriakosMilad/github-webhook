package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type PushEvent struct {
	Ref  string `json:"ref"`
	Repo struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
	}
}

func main() {
	err := loadEnvVars()
	if err != nil {
		log.Fatal(err)
		return
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

func loadEnvVars() error {
	// open the .env file
	file, err := os.Open(".env")
	if err != nil {
		return fmt.Errorf("could not open .env file: %v", err)
	}
	defer file.Close()

	// read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// ignore empty lines or comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// split key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // skip invalid lines
		}

		// set the environment variable
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		err := os.Setenv(key, value)
		if err != nil {
			return fmt.Errorf("could not set environment variable %s: %v", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("could not read .env file: %v", err)
	}
	return nil
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// get needed environment variables
	githubWebhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	branchName := os.Getenv("BRANCH_NAME")
	repoFullName := os.Getenv("REPO_FULL_NAME")
	shellPath := os.Getenv("SHELL_PATH")

	// read the body of the request (webhook payload)
	payload := make([]byte, r.ContentLength)
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// verify the request signature
	isValid := verifySignature(payload, githubWebhookSecret, r.Header.Get("X-Hub-Signature-256"))
	if !isValid {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// parse the GitHub push event
	var pushEvent PushEvent
	err = json.Unmarshal(payload, &pushEvent)
	if err != nil {
		http.Error(w, "Error parsing webhook payload", http.StatusBadRequest)
		return
	}

	// check if push event is for the correct repository
	if pushEvent.Repo.FullName != repoFullName {
		// ignore pushes to other repositories
		log.Printf("Ignored push to repository: %s", pushEvent.Repo.FullName)
		fmt.Fprintf(w, "Ignored push to repository: %s", pushEvent.Repo.FullName)

		// return 200 OK status
		w.WriteHeader(http.StatusOK)
		return
	}

	// check if the push is to the correct branch
	if strings.HasSuffix(pushEvent.Ref, branchName) == false {
		// ignore pushes to other branches
		log.Printf("Ignored push to branch: %s", pushEvent.Ref)
		fmt.Fprintf(w, "Ignored push to branch: %s", pushEvent.Ref)

		// return 200 OK status
		w.WriteHeader(http.StatusOK)
		return
	}

	// execute the shell script
	cmd := exec.Command(shellPath)
	err = cmd.Run()
	if err != nil {
		http.Error(w, "Error executing shell script", http.StatusInternalServerError)
		return
	}

	// return 200 OK status
	w.WriteHeader(http.StatusOK)
}

func verifySignature(payload []byte, githubWebhookSecret, signature string) bool {
	// compute HMAC-SHA256 hash of the payload using the secret key
	mac := hmac.New(sha256.New, []byte(githubWebhookSecret))
	mac.Write(payload)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// compare the computed hash with the signature sent by GitHub
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}
