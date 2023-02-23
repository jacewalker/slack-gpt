// running on PID 4128995
// nohup ./myexecutable &
// kill <pid>

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
	chat := AskGPT{}
	r := gin.Default()

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	chat.APIKey = os.Getenv("OPENAI_AUTHKEY")

	r.POST("/api/slack", func(c *gin.Context) {
		c.Status(200)

		log.Println("RECEIVED: New Post Request")
		// slackObj, challenge := slack.ParsePostRequest(c)
		chat.SlackObject, chat.SlackChallenge = slack.ParsePostRequest(c)

		// Handle the Challenge Response
		if chat.SlackChallenge != "" {
			go slack.RespondToChallenge(&chat.SlackChallenge, c)
		}

		chat.EventType = chat.SlackObject.Event.Type

		switch chat.EventType {
		case "app_mention":
			go processResponse(&chat)
		case "member_joined_channel": // this is inactive
			user := chat.SlackObject.Event.User
			prompt := fmt.Sprintf("Write a short welcome message to the new user, %s", user)
			go openai.MakePrompt(prompt, &chat.APIKey, &chat.SlackObject)
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

		// Check if the status page has an ongoing outage
		if body["status"].(string) == "outage" {
			fmt.Printf("OpenAI is currently experiencing an outage: %s\n", body["message"].(string))
			outageEmailMessage := fmt.Sprintf("OpenAI is currently experiencing an outage: %s\n", body["message"].(string))
			go email.SendEmail("jacewalker@me.com", outageEmailMessage)
		}

		c.Status(http.StatusOK)
	})

	r.Run(":8080")
}

func processResponse(chat *AskGPT) {
	chat.Prompt = chat.SlackObject.Event.Blocks[0].Elements1[0].Elements2[1].UserText
	historyMap, _ := dbops.LookupFromDatabase(chat.SlackObject.Event.ThreadTS)
	history := openai.CreateHistoricPrompt(historyMap, chat.Prompt)
	fmt.Println("History String:\n", history)

	completion := openai.MakePrompt(history, &chat.APIKey, &chat.SlackObject)
	dbops.SaveToDatabase(chat.SlackObject, &completion)

	err := slack.SendMessage(&completion, &chat.SlackObject)
	if err != nil {
		log.Println("[WARNING] Unable to send Slack Message:", err)
		emailMessage := fmt.Sprintf("Failed to respond to Slack message with error: %s\n", err)
		email.SendEmail("jacewalker@me.com", emailMessage)
	}

	// respType := openai.CheckPromptType(prompt, apiKey)
	// fmt.Println("Response Type is", respType)

	// switch respType {
	// case "0":
	// 	go openai.MakePrompt(prompt, apiKey, slackObj)
	// case "1":
	// 	fmt.Println("!!!!!!!!! Logging a ticket...")
	// 	go email.SendEmail("jacewalker@me.com")
	// 	go slack.SendMessage("Ok, I have logged a ticket for this!", slackObj)
	// default:
	// 	go openai.MakePrompt(prompt, apiKey, slackObj)
	// }
}
