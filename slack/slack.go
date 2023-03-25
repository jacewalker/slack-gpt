package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type SlackChallenge struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}

type SlackEvent struct {
	Challenge string `json:"challenge"`
	Token     string `json:"token"`
	TeamID    string `json:"team_id"`
	ApiAppID  string `json:"api_app_id"`
	Event     struct {
		ClientMsgID string `json:"client_msg_id"`
		Type        string `json:"type"`
		Text        string `json:"text"`
		User        string `json:"user"`
		TS          string `json:"ts"`
		Blocks      []struct {
			Type      string `json:"type"`
			BlockID   string `json:"block_id"`
			Elements1 []struct {
				Type      string `json:"type"`
				Elements2 []struct {
					Type     string `json:"type"`
					UserID   string `json:"user_id,omitempty"`
					UserText string `json:"text,omitempty"`
				} `json:"elements"`
			} `json:"elements"`
		} `json:"blocks"`
		Team     string `json:"team"`
		ThreadTS string `json:"thread_ts"`
		Channel  string `json:"channel"`
		EventTS  string `json:"event_ts"`
	} `json:"event"`
	Type           string `json:"type"`
	EventID        string `json:"event_id"`
	EventTime      int    `json:"event_time"`
	Authorizations []struct {
		EnterpriseID        string `json:"enterprise_id,omitempty"`
		TeamID              string `json:"team_id"`
		UserID              string `json:"user_id"`
		IsBot               bool   `json:"is_bot"`
		IsEnterpriseInstall bool   `json:"is_enterprise_install"`
	} `json:"authorizations"`
	IsExtSharedChannel bool   `json:"is_ext_shared_channel"`
	EventContext       string `json:"event_context"`
}

func RespondToChallenge(challenge *string, c *gin.Context) {
	jsonStr := []byte(`{"challenge":"` + *challenge + `"}`)
	log.Info().Msgf("Challenge Response:", string(jsonStr))
	c.JSON(http.StatusOK, string(jsonStr))
}

func ParsePostRequest(c *gin.Context) (object SlackEvent, challenge string) {
	var result SlackEvent
	bodyAsByteArray, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(bodyAsByteArray, &result); err != nil {
		log.Info().Msg(err.Error())
	}

	// Write the JSON request to a file
	filename := fmt.Sprintf("postrequests/%d.json", time.Now().UnixNano())
	if err := ioutil.WriteFile(filename, bodyAsByteArray, 0644); err != nil {
		log.Info().Msgf("Error writing JSON request to file:", err)
	}

	if result.Challenge != "" {
		log.Info().Msgf("[INFO] Received Challenge Request:", result.Challenge)
		return result, result.Challenge
	} else {
		log.Info().Msgf("[INFO] Received POST Request from", c.Request.UserAgent())
		log.Info().Msgf("[INFO] Client IP Address:", c.ClientIP())
		// log.Info().Msgf("[INFO] User Provided Text:", result.Event.Blocks[0].Elements1[0].Elements2[1].UserText)
		log.Info().Msgf("[INFO] User Provided Text:", result.Event.Text)
		return result, ""
	}
}

func textToJSONString(text, thread string) ([]byte, error) {
	textJSON := struct {
		Type   string `json:"type"`
		Text   string `json:"text"`
		Thread string `json:"thread_ts"`
	}{
		Type:   "mrkdwn",
		Text:   text,
		Thread: thread,
	}

	textJSONBytes, err := json.Marshal(textJSON)
	if err != nil {
		return []byte{}, err
	}

	return textJSONBytes, nil
}

func SendMessage(response *string, object *SlackEvent) error {
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	log.Info().Msgf("Error loading .env file")
	// }
	webhookUrl := os.Getenv("SLACK_AICHATBOT_URL")

	jsonStr, err := textToJSONString(*response, object.Event.TS)
	if err != nil {
		log.Info().Msgf("[WARNING] Failure converting outgoing message to JSON:", err)
	}

	resp, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil || resp.StatusCode != 200 {
		log.Info().Msgf("Error not nil or resp not 200, resp:", resp)
		return err
	} else {
		return nil
	}
}
