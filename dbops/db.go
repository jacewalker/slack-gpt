package dbops

import (
	"errors"
	"os"

	"github.com/jacewalker/slack-gpt/slack"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	Prompt         string
	Completion     string
	User           string
	TimeStamp      string
	ThreadTS       string
	Team           string
	Channel        string
	EventTimeStamp string
}

func InitDatabase() *gorm.DB {
	dbName := "files.db"

	if _, err := os.Stat(dbName); errors.Is(err, os.ErrNotExist) {
		log.Info().Msgf(dbName, "doesn't exist. Creating file...")
		if _, err := os.Create(dbName); err != nil {
			log.Error().Msgf("Unable to create files.db:", err)
		}
	} else {
		log.Info().Msgf(dbName, "already exists. Continuing using existing database.")
	}

	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Error().Msgf("Unable to connect to the database:", err)
	} else {
		log.Info().Msgf("Database Initialised Successfully")
	}
	err = db.AutoMigrate(&Message{})
	if err != nil {
		log.Error().Msgf("Unable to complete database migration:", err)
	} else {
		log.Info().Msgf("Database migration completed.")
	}
	return db
}

func SaveToDatabase(event *slack.SlackEvent, completion *string, db *gorm.DB) {
	log.Info().Msg("Saving message transaction to the database.")

	db.Create(&Message{
		Prompt:         event.Event.Blocks[0].Elements1[0].Elements2[1].UserText,
		Completion:     *completion,
		User:           event.Event.Blocks[0].Elements1[0].Elements2[0].UserID,
		TimeStamp:      event.Event.TS,
		ThreadTS:       event.Event.ThreadTS,
		Team:           event.Event.Team,
		Channel:        event.Event.Channel,
		EventTimeStamp: event.Event.EventTS,
	})

	LookupFromDatabase(&event.Event.ThreadTS, db)
}

func LookupFromDatabase(threadTS *string, db *gorm.DB) (messages map[string]string, err error) {
	log.Info().Msg("Looking up further transactions from the database.")

	var msgs []Message
	var history = make(map[string]string)

	if *threadTS == "" {
		return nil, errors.New("empty thread_ts")
	}

	db.Where("thread_ts = ? OR time_stamp = ?", threadTS, threadTS).Order("created_at asc").Find(&msgs)

	for _, msg := range msgs {
		history[msg.Prompt] = msg.Completion
	}

	return history, nil
}
