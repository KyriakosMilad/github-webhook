package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type PushEvent struct {
	Repo struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file: " + err.Error())
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

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// get needed environment variables
	githubWebhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	//branchName := os.Getenv("BRANCH_NAME")
	repoFullName := os.Getenv("REPO_FULL_NAME")
	shellPath := os.Getenv("SHELL_PATH")

	// read the body of the request (webhook payload)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// verify the request signature
	signature := strings.Split(r.Header.Get("X-Hub-Signature-256"), "=")
	if len(signature) != 2 {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}
	isValid := verifySignature(payload, githubWebhookSecret, signature[1])
	if !isValid {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// parse the request body
	pushEvent := &PushEvent{}
	err = json.Unmarshal(payload, pushEvent)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid payload: %s", err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// check if push event is for the correct repository
	if pushEvent.Repo.FullName != repoFullName {
		// ignore pushes to other repositories
		log.Printf("Ignored push to repository: %s", pushEvent.Repo.FullName)
		fmt.Fprintf(w, "Ignored push to repository: %s", pushEvent.Repo.FullName)

		// return 200 OK status
		w.WriteHeader(http.StatusOK)
		return
	}

	// for some reason, the branch name is not included in the payload, comment this check for now
	// check if the push is to the correct branch
	//if strings.HasSuffix(pushEvent.Ref, branchName) == false {
	//	// ignore pushes to other branches
	//	log.Printf("Ignored push to branch: %s", pushEvent.Ref)
	//	fmt.Fprintf(w, "Ignored push to branch: %s", pushEvent.Ref)
	//
	//	// return 200 OK status
	//	w.WriteHeader(http.StatusOK)
	//	return
	//}

	// Create buffers to capture both stdout and stderr
	var outBuf, errBuf bytes.Buffer

	// execute the shell script
	cmd := exec.Command("bash", shellPath)
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
