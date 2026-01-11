package logic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/navyaalva/sbf-os/internal/db"
)

// HTTP client with timeout for external API calls
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// Subtask structure for JSON
type Subtask struct {
	Title  string `json:"title"`
	IsDone bool   `json:"is_done"`
}

// Gemini Request Structure
type GeminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

// Gemini Response Structure
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// 1. GenerateSubtasks (REAL AI VERSION)
func GenerateSubtasks(taskTitle, description string) []byte {
	apiKey := os.Getenv("GEMINI_API_KEY")

	// Fallback to simulation if no key (so app doesn't crash)
	if apiKey == "" {
		fmt.Println("⚠️ AI Warning: No GEMINI_API_KEY found. Using simulation.")
		return generateSimulatedSubtasks(taskTitle)
	}

	// The Prompt
	prompt := fmt.Sprintf(`
		You are an expert event planner. 
		Break down this task into 3-5 concrete, actionable steps.
		Task: "%s"
		Context: "%s"
		
		Return ONLY raw JSON in this format: 
		[{"title": "Step 1", "is_done": false}, {"title": "Step 2", "is_done": false}]
	`, taskTitle, description)

	// Prepare Request
	reqBody, _ := json.Marshal(GeminiRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{Parts: []struct {
				Text string `json:"text"`
			}{{Text: prompt}}},
		},
	})

	// UPDATED MODEL: gemini-2.5-flash
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey
	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("❌ AI Network Error:", err)
		return generateSimulatedSubtasks(taskTitle)
	}
	defer resp.Body.Close()

	// Parse Response
	body, _ := io.ReadAll(resp.Body)

	// --- DEBUG PRINT ---
	fmt.Println("🤖 AI Response from Google:", string(body))
	// -------------------

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		fmt.Println("❌ AI JSON Error:", err)
		return generateSimulatedSubtasks(taskTitle)
	}

	if len(geminiResp.Candidates) == 0 {
		fmt.Println("❌ AI Error: No candidates returned.")
		return generateSimulatedSubtasks(taskTitle)
	}

	text := geminiResp.Candidates[0].Content.Parts[0].Text

	// --- NEW BULLETPROOF CLEANING LOGIC ---
	// Instead of guessing markdown tags, we just find the JSON array brackets.
	start := strings.Index(text, "[")
	end := strings.LastIndex(text, "]")

	if start != -1 && end != -1 && end > start {
		text = text[start : end+1]
	} else {
		// If we can't find brackets, something is wrong with the AI output
		fmt.Println("❌ AI Error: Could not find JSON array in response.")
		return generateSimulatedSubtasks(taskTitle)
	}
	// ---------------------------------------

	return []byte(text)
}

// Fallback logic
func generateSimulatedSubtasks(title string) []byte {
	lowerTitle := strings.ToLower(title)
	var steps []Subtask

	if strings.Contains(lowerTitle, "venue") {
		steps = []Subtask{
			{Title: "Research capacity options", IsDone: false},
			{Title: "Schedule site visits", IsDone: false},
			{Title: "Review contract terms", IsDone: false},
		}
	} else if strings.Contains(lowerTitle, "vendor") {
		steps = []Subtask{
			{Title: "Create application form", IsDone: false},
			{Title: "Email past vendors", IsDone: false},
			{Title: "Collect payments", IsDone: false},
		}
	} else {
		steps = []Subtask{
			{Title: "Draft initial plan (AI Offline)", IsDone: false},
			{Title: "Review with team", IsDone: false},
			{Title: "Execute", IsDone: false},
		}
	}

	b, _ := json.Marshal(steps)
	return b
}

// 2. CheckFollowUps
func CheckFollowUps(tasks []db.Task) []string {
	var reminders []string

	for _, t := range tasks {
		if !t.DueDate.Valid || t.AssigneeText.String == "" {
			continue
		}

		daysUntil := int(time.Until(t.DueDate.Time).Hours() / 24)
		assignee := t.AssigneeText.String

		if daysUntil < 0 {
			reminders = append(reminders, fmt.Sprintf("🚨 <strong>%s</strong> is overdue on '%s'", assignee, t.Title))
		} else if daysUntil <= 2 {
			reminders = append(reminders, fmt.Sprintf("👉 Nudge <strong>%s</strong> about '%s' (Due in %d days)", assignee, t.Title, daysUntil))
		}
	}
	return reminders
}
