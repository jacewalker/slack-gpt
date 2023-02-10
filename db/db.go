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

func SaveToDatabase(event slack.SlackEvent, completion string) {
	db := InitDatabase()
	db.Create(&Message{
		Prompt:         event.Event.Blocks[0].Elements1[0].Elements2[1].UserText,
		Completion:     completion,
		User:           event.Event.Blocks[0].Elements1[0].Elements2[0].UserID,
		TimeStamp:      event.Event.TS,
		ThreadTS:       event.Event.EventTS,
		Team:           event.Event.Team,
		Channel:        event.Event.Channel,
		EventTimeStamp: event.Event.EventTS,
	})
}

// func LookupFromDatabase(threadTS string) (message string) {
// 	db := InitDatabase()
// 	var msg Message

// 	db.Where("threadts <> ?", threadTS).Find(&msg)
// 	if err := db.Where(&msg, "threadts = ?", threadTS).Error; errors.Is(err, gorm.ErrRecordNotFound) {
// 		fmt.Printf("Error: ThreadTS %s not found.", threadTS)
// 		return "", gorm.ErrRecordNotFound
// 	} else if db.Error != nil {
// 		fmt.Println("Error on database lookup:", db.Error)
// 		return "", gorm.ErrRecordNotFound
// 	} else {
// 		log.Println("Thread History found:", msg.ThreadTS)
// 		return file.Name, nil
// 	}
// }

// func DeleteFromDatabase(key string) {
// 	var message Message
// 	db := InitDatabase()
// 	db.Delete(&Message, "key = ?", key)
// }
