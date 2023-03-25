package openaistatus

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jacewalker/slack-gpt/email"
	"github.com/rs/zerolog/log"
)

type Incident struct {
	Backfilled                    bool        `json:"backfilled"`
	CreatedAt                     string      `json:"created_at"`
	Impact                        string      `json:"impact"`
	ImpactOverride                interface{} `json:"impact_override"`
	MonitoringAt                  string      `json:"monitoring_at"`
	PostmortemBody                interface{} `json:"postmortem_body"`
	PostmortemBodyLastUpdatedAt   interface{} `json:"postmortem_body_last_updated_at"`
	PostmortemIgnored             bool        `json:"postmortem_ignored"`
	PostmortemNotifiedSubscribers bool        `json:"postmortem_notified_subscribers"`
	PostmortemNotifiedTwitter     bool        `json:"postmortem_notified_twitter"`
	PostmortemPublishedAt         interface{} `json:"postmortem_published_at"`
	ResolvedAt                    interface{} `json:"resolved_at"`
	ScheduledAutoTransition       bool        `json:"scheduled_auto_transition"`
	ScheduledFor                  interface{} `json:"scheduled_for"`
	ScheduledRemindPrior          bool        `json:"scheduled_remind_prior"`
	ScheduledRemindedAt           interface{} `json:"scheduled_reminded_at"`
	ScheduledUntil                interface{} `json:"scheduled_until"`
	Shortlink                     string      `json:"shortlink"`
	Status                        string      `json:"status"`
	UpdatedAt                     string      `json:"updated_at"`
	ID                            string      `json:"id"`
	OrganizationID                string      `json:"organization_id"`
	IncidentUpdates               []struct {
		Body               string      `json:"body"`
		CreatedAt          string      `json:"created_at"`
		DisplayAt          string      `json:"display_at"`
		Status             string      `json:"status"`
		TwitterUpdatedAt   interface{} `json:"twitter_updated_at"`
		UpdatedAt          string      `json:"updated_at"`
		WantsTwitterUpdate bool        `json:"wants_twitter_update"`
		ID                 string      `json:"id"`
		IncidentID         string      `json:"incident_id"`
	} `json:"incident_updates"`
	Name string `json:"name"`
}

type Page struct {
	ID                string `json:"id"`
	StatusIndicator   string `json:"status_indicator"`
	StatusDescription string `json:"status_description"`
}

type Meta struct {
	Unsubscribe   string `json:"unsubscribe"`
	Documentation string `json:"documentation"`
}

type PostRequest struct {
	Meta     Meta     `json:"meta"`
	Page     Page     `json:"page"`
	Incident Incident `json:"incident"`
}

func (pr *PostRequest) NotifyOutage() {
	log.Info().Msgf("OpenAI is currently experiencing an outage: %s\n", pr.Incident.Name)
	outageEmailMessage := fmt.Sprintf(`OpenAI is currently experiencing an outage: %s
	Additional information can be found here: %s`, pr.Incident.Name, pr.Incident.Shortlink)
	go email.SendEmail("jacewalker@me.com", outageEmailMessage)
}

func StatusUpdate(c *gin.Context) {
	var postRequest PostRequest
	if err := c.BindJSON(&postRequest); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	postRequest.NotifyOutage()

	c.Status(http.StatusOK)
}
