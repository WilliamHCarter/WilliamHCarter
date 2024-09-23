package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// =========================== Queries ===========================
func GetTotalCommits(username string) (int, error) {
	logger.Printf("Fetching commit count for user: %s\n", username)

	url := fmt.Sprintf("https://api.github.com/search/commits?q=author:%s", username)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Printf("Error creating request: %v\n", err)
		return 0, err
	}

	req.Header.Set("User-Agent", "GetCommitsAgent")
	req.Header.Set("Accept", "application/vnd.github.cloak-preview")

	logger.Println("Sending request to GitHub API...")
	resp, err := client.Do(req)
	if err != nil {
		logger.Printf("Error sending request: %v\n", err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Printf("API request failed with status code: %d\n", resp.StatusCode)
		return 0, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var result struct {
		TotalCount int `json:"total_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Printf("Error decoding response: %v\n", err)
		return 0, err
	}

	logger.Printf("Total commits counted: %d\n", result.TotalCount)
	return result.TotalCount, nil
}
