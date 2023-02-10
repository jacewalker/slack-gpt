// running on PID 1329858
// nohup ./myexecutable &
// disown <pid>

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jacewalker/slack-gpt/openai"
	"github.com/jacewalker/slack-gpt/slack"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("OPENAI_AUTHKEY")

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
			go processResponse(apiKey, slackObj)
		case "member_joined_channel":
			// c.Status(http.StatusOK)
			user := slackObj.Event.User
			prompt := fmt.Sprintf("Write a short welcome message to the new user, %s", user)
			go openai.MakePrompt(prompt, apiKey, slackObj)
		default:
			fmt.Println("Unknown request! Maybe a challenge?")
		}
	})

	r.Run(":8080")
}

func processResponse(apiKey string, slackObj slack.SlackEvent) {
	prompt := slackObj.Event.Blocks[0].Elements1[0].Elements2[1].UserText
	respType := openai.CheckPromptType(prompt, apiKey)
	fmt.Println("Response Type is", respType)

	go openai.MakePrompt(prompt, apiKey, slackObj)

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
