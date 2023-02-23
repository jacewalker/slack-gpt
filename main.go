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
	SlackChallenge slack.SlackChallenge
	EventType      string
}

func main() {
	chat := AskGPT{}

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	chat.APIKey = os.Getenv("OPENAI_AUTHKEY")

	r := gin.Default()

	r.POST("/api/slack", func(c *gin.Context) {
		c.Status(200)

		log.Println("RECEIVED: New Post Request")
		slackObj, challenge := slack.ParsePostRequest(c)

		// Handle the Challenge Response
		if challenge != "" {
			slack.RespondToChallenge(challenge, c)
		}

		event := slackObj.Event.Type

		switch event {
		case "app_mention":
			go processResponse(&chat.APIKey, slackObj)
		case "member_joined_channel": // this is inactive
			user := slackObj.Event.User
			prompt := fmt.Sprintf("Write a short welcome message to the new user, %s", user)
			go openai.MakePrompt(prompt, &chat.APIKey, slackObj)
		default:
			fmt.Println("Unknown request! Maybe a challenge?")
		}
	})

	r.POST("/api/openai-status", func(c *gin.Context) {
		// Parse the request body
		var body map[string]interface{}
		if err := c.BindJSON(&body); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		// Check if the status page has an ongoing outage
		if body["status"].(string) == "outage" {
			// Print the outage message to the console
			fmt.Printf("OpenAI is currently experiencing an outage: %s\n", body["message"].(string))
			outageEmailMessage := fmt.Sprintf("OpenAI is currently experiencing an outage: %s\n", body["message"].(string))
			email.SendEmail("jacewalker@me.com", outageEmailMessage)
		}

		c.Status(http.StatusOK)
	})

	r.Run(":8080")
}

func processResponse(apiKey *string, slackObj slack.SlackEvent) {
	prompt := slackObj.Event.Blocks[0].Elements1[0].Elements2[1].UserText
	// respType := openai.CheckPromptType(prompt, apiKey)
	// fmt.Println("Response Type is", respType)

	historyMap, _ := dbops.LookupFromDatabase(slackObj.Event.ThreadTS)
	history := openai.CreateHistoricPrompt(historyMap, prompt)
	fmt.Println("History String:\n", history)

	completion := openai.MakePrompt(history, apiKey, slackObj)
	dbops.SaveToDatabase(slackObj, completion)

	err := slack.SendMessage(completion, slackObj)
	if err != nil {
		log.Println("[WARNING] Unable to send Slack Message:", err)
	}

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
