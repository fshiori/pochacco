package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	To   struct {
		Name string `json:"name"`
	} `json:"to"`
}

type TransitionsResponse struct {
	Transitions []Transition `json:"transitions"`
}

type TransitionRequest struct {
	Transition struct {
		ID string `json:"id"`
	} `json:"transition"`
}

type IssueStatus struct {
	Fields struct {
		Status struct {
			Name string `json:"name"`
		} `json:"status"`
	} `json:"fields"`
}

func main() {
	// 定義命令行參數
	issueKey := flag.String("issue", "", "Jira issue key (e.g., PRD-3936)")
	transitionID := flag.String("transition", "", "Transition ID to change status")
	flag.Parse()

	// 檢查必要的參數
	if *issueKey == "" {
		log.Fatal("Please provide issue key using -issue flag")
	}

	// 載入 .env 文件，如果失敗則繼續執行
	_ = godotenv.Load()

	// 從環境變數獲取設定
	siteURL := os.Getenv("JIRA_SITE")
	username := os.Getenv("JIRA_USER")
	token := os.Getenv("JIRA_TOKEN")

	// 檢查必要的環境變數是否存在
	if siteURL == "" || username == "" || token == "" {
		log.Fatal("Missing required environment variables: JIRA_SITE, JIRA_USER, or JIRA_TOKEN")
	}

	// 建立基本的 HTTP client
	client := &http.Client{}

	// 如果提供了 transition ID，執行狀態轉換
	if *transitionID != "" {
		performTransition(client, siteURL, username, token, *issueKey, *transitionID)
		return
	}

	// 否則顯示可用的轉換狀態
	listTransitions(client, siteURL, username, token, *issueKey)
}

func listTransitions(client *http.Client, siteURL, username, token, issueKey string) {
	transitionsURL := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", siteURL, issueKey)

	req, err := http.NewRequest("GET", transitionsURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(username, token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var transitions TransitionsResponse
	err = json.Unmarshal(body, &transitions)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Available transitions for issue %s:\n", issueKey)
	for _, t := range transitions.Transitions {
		fmt.Printf("ID: %s, Name: %s, To Status: %s\n", t.ID, t.Name, t.To.Name)
	}
}

func performTransition(client *http.Client, siteURL, username, token, issueKey, transitionID string) {
	transitionsURL := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", siteURL, issueKey)

	// 準備請求體
	transitionReq := TransitionRequest{}
	transitionReq.Transition.ID = transitionID

	jsonData, err := json.Marshal(transitionReq)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", transitionsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(username, token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		fmt.Printf("Successfully transitioned issue %s to status with ID %s\n", issueKey, transitionID)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Failed to transition issue. Status: %d, Response: %s\n", resp.StatusCode, string(body))
	}
}
