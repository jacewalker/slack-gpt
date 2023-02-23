package dbops

import (
	"errors"
	"fmt"
	"os"

	"github.com/jacewalker/slack-gpt/slack"
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
	// check if the db has already been created/initialised before proceeding
	dbName := "files.db"

	if _, err := os.Stat(dbName); errors.Is(err, os.ErrNotExist) {
		fmt.Println(dbName, "doesn't exist. Creating file...")
		if _, err := os.Create(dbName); err != nil {
			fmt.Println("Unable to create files.db:", err)
		}
	} else {
		fmt.Println(dbName, "already exists. Continuing using existing database.")
	}

	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		fmt.Println("Unable to connect to the database:", err)
	} else {
		fmt.Println("Database Initialised Successfully")
	}
	err = db.AutoMigrate(&Message{})
	if err != nil {
		fmt.Println("Unable to complete database migration:", err)
	} else {
		fmt.Println("Database migration completed.")
	}
	return db
}

func SaveToDatabase(event slack.SlackEvent, completion *string) {
	db := InitDatabase()
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

	LookupFromDatabase(event.Event.ThreadTS)
}

func LookupFromDatabase(threadTS string) (messages map[string]string, err error) {
	db := InitDatabase()
	var msgs []Message
	var history = make(map[string]string)

	if threadTS == "" {
		return nil, errors.New("empty thread_ts")
	}

	db.Where("thread_ts = ? OR time_stamp = ?", threadTS, threadTS).Order("created_at asc").Find(&msgs)

	for _, msg := range msgs {
		history[msg.Prompt] = msg.Completion
	}

	return history, nil
}

// func DeleteFromDatabase(key string) {
// 	var message Message
// 	db := InitDatabase()
// 	db.Delete(&Message, "key = ?", key)
// }
