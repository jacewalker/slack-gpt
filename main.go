// running on PID 73599
// nohup ./myexecutable &
// kill <pid>

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jacewalker/slack-gpt/dbops"
	"github.com/jacewalker/slack-gpt/email"
	"github.com/jacewalker/slack-gpt/openai"
	openaistatus "github.com/jacewalker/slack-gpt/openai-status"
	"github.com/jacewalker/slack-gpt/slack"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type AskGPT struct {
	APIKey         string
	SlackObject    slack.SlackEvent
	SlackChallenge string
	EventType      string
	Prompt         string
	Database       *gorm.DB
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	r := gin.Default()

	if err := godotenv.Load(".env"); err != nil {
		log.Info().Msg(".env file not found. Using environment variables.")
	}

	chat := AskGPT{}
	chat.APIKey = os.Getenv("OPENAI_AUTHKEY")
	chat.Database = dbops.InitDatabase()

	r.POST("/api/slack", func(c *gin.Context) {
		c.Status(200)

		if ct := c.ContentType(); ct != "application/json" {
			log.Error().Msgf("Content Type Error: %s", ct)
			c.AbortWithStatusJSON(415, gin.H{
				"error": "Invalid Content Type, expected application/json",
			})
			return
		}

		chat.SlackObject, chat.SlackChallenge = slack.ParsePostRequest(c)

		if chat.SlackChallenge != "" {
			slack.RespondToChallenge(&chat.SlackChallenge, c)
		}

		switch chat.SlackObject.Event.Type {
		case "app_mention":
			go processResponse(&chat)
		default:
			log.Debug().Msg("Unknown request... Maybe a challenge?")
			log.Debug().Msg(chat.SlackObject.Event.Type)
		}
	})

	r.POST("/api/openai-status", openaistatus.StatusUpdate)

	r.Run(":80")
}

func processResponse(chat *AskGPT) {
	// chat.Prompt = chat.SlackObject.Event.Blocks[0].Elements1[0].Elements2[1].UserText
	log.Info().Msg("Received App Mention from Slack")
	chat.Prompt = chat.SlackObject.Event.Text

	var respType string

	if strings.Contains(chat.Prompt, "ticket") {
		log.Info().Msg("Slack Message contains 'ticket' trigger word.")
		respType = openai.CheckPromptType(chat.Prompt, &chat.APIKey)
	}

	switch {
	case respType == "" || strings.Contains(respType, "0"):
		historyMap, _ := dbops.LookupFromDatabase(&chat.SlackObject.Event.ThreadTS, chat.Database)
		history := openai.CreateHistoricChatPrompt(historyMap, chat.Prompt)

		completion := openai.MakeChatPrompt(chat.Prompt, &chat.APIKey, history)
		go dbops.SaveToDatabase(&chat.SlackObject, &completion, chat.Database)

		err := slack.SendMessage(&completion, &chat.SlackObject)
		log.Info().Msg("Sent response to the Slack thread")
		if err != nil {
			log.Error().Msgf("[WARNING] Unable to send Slack Message:", err)
			emailMessage := fmt.Sprintf("Failed to respond to Slack message with error: %s\n", err)
			email.SendEmail("jacewalker@me.com", emailMessage)
		}
	case strings.Contains(respType, "1"):
		log.Info().Msg("1: Logging a ticket")
		success := email.SendEmail("jacewalker@me.com", "New Ticket Logged!")
		var response string
		if success {
			response = "[You've found a feature that's coming soon!]"
		} else {
			response = "I tried logging a ticket but had an issue sending the email to support@ottoit.com.au. Log it manually for now until I get myself sorted."
		}
		go slack.SendMessage(&response, &chat.SlackObject)
	default:
		log.Error().Msg("Unsure")
	}

}
