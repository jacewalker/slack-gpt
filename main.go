// running on PID 2289667
// nohup ./myexecutable &
// kill <pid>

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jacewalker/slack-gpt/dbops"
	"github.com/jacewalker/slack-gpt/email"
	"github.com/jacewalker/slack-gpt/openai"
	"github.com/jacewalker/slack-gpt/slack"
	"github.com/joho/godotenv"
)

type AskGPT struct {
	APIKey         string
	SlackObject    slack.SlackEvent
	SlackChallenge string
	EventType      string
	Prompt         string
}

func main() {
	r := gin.Default()

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	chat := AskGPT{}
	chat.APIKey = os.Getenv("OPENAI_AUTHKEY")

	r.POST("/api/slack", func(c *gin.Context) {
		c.Status(200)
		chat.SlackObject, chat.SlackChallenge = slack.ParsePostRequest(c)

		if chat.SlackChallenge != "" {
			slack.RespondToChallenge(&chat.SlackChallenge, c)
		}

		switch chat.SlackObject.Event.Type {
		case "app_mention":
			go processResponse(&chat)
		default:
			fmt.Println("Unknown request! Maybe a challenge?")
		}
	})

	r.POST("/api/openai-status", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.BindJSON(&body); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		fmt.Printf("OpenAI is currently experiencing an outage: %s\n", body["message"].(string))
		outageEmailMessage := fmt.Sprintf("OpenAI is currently experiencing an outage: %s\n", body["message"].(string))
		go email.SendEmail("jacewalker@me.com", outageEmailMessage)

		c.Status(http.StatusOK)
	})

	r.Run(":8080")
}

func processResponse(chat *AskGPT) {
	chat.Prompt = chat.SlackObject.Event.Blocks[0].Elements1[0].Elements2[1].UserText

	var respType string

	if strings.Contains(chat.Prompt, "ticket") {
		respType = openai.CheckPromptType(chat.Prompt, &chat.APIKey)
	}

	switch {
	case respType == "" || strings.Contains(respType, "0"):
		historyMap, _ := dbops.LookupFromDatabase(chat.SlackObject.Event.ThreadTS)
		history := openai.CreateHistoricChatPrompt(historyMap, chat.Prompt)
		fmt.Println("History String:\n", history)

		completion := openai.MakeChatPrompt(chat.Prompt, &chat.APIKey, history)
		dbops.SaveToDatabase(chat.SlackObject, &completion)

		err := slack.SendMessage(&completion, &chat.SlackObject)
		if err != nil {
			log.Println("[WARNING] Unable to send Slack Message:", err)
			emailMessage := fmt.Sprintf("Failed to respond to Slack message with error: %s\n", err)
			email.SendEmail("jacewalker@me.com", emailMessage)
		}
	case strings.Contains(respType, "1"):
		fmt.Println("1: Logging a ticket")
		success := email.SendEmail("jacewalker@me.com", "New Ticket Logged!")
		var response string
		if success {
			response = "[You've found a feature that's coming soon!]"
		} else {
			response = "I tried logging a ticket but had an issue sending the email to support@ottoit.com.au. Log it manually for now until I get myself sorted."
		}
		go slack.SendMessage(&response, &chat.SlackObject)
	default:
		fmt.Println("Unsure")
	}

}
