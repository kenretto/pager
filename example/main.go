package main

import (
	"encoding/json"
	"github.com/kenretto/pager"
	"github.com/kenretto/pager/driver"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"os"
	"time"
)

type user struct {
	gorm.Model
	Nickname string `json:"nickname" gorm:"column:nickname;type:varchar(16)"`
	Age      uint8  `json:"age" gorm:"column:age;"`
}

// TableName table name
func (user) TableName() string {
	return "members"
}

var u user

var users = []user{
	{Nickname: "a", Age: 1},
	{Nickname: "b", Age: 2},
	{Nickname: "c", Age: 3},
	{Nickname: "d", Age: 4},
	{Nickname: "e", Age: 5},
	{Nickname: "f", Age: 6},
	{Nickname: "g", Age: 7},
	{Nickname: "h", Age: 8},
	{Nickname: "i", Age: 9},
	{Nickname: "j", Age: 10},
	{Nickname: "k", Age: 11},
	{Nickname: "l", Age: 12},
	{Nickname: "m", Age: 13},
	{Nickname: "n", Age: 14},
	{Nickname: "o", Age: 15},
	{Nickname: "p", Age: 16},
	{Nickname: "q", Age: 17},
	{Nickname: "r", Age: 18},
	{Nickname: "s", Age: 19},
	{Nickname: "t", Age: 20},
	{Nickname: "u", Age: 21},
	{Nickname: "v", Age: 22},
	{Nickname: "w", Age: 23},
	{Nickname: "x", Age: 24},
	{Nickname: "y", Age: 25},
	{Nickname: "z", Age: 26},
}

func main() {
	var db, err = gorm.Open(sqlite.Open("file:pager?mode=memory&cache=shared&_fk=1"), &gorm.Config{PrepareStmt: true, Logger: logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      false,
		},
	)})
	if err != nil {
		log.Fatalln(err)
	}
	_ = db.AutoMigrate(&u)
	db.Model(&u).Create(&users)

	http.HandleFunc("/member", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(pager.New(request, driver.NewGormDriver(db)).
			SetPaginationField("ID").SetIndex(u.TableName()).
			Find(u).Result())
		_, _ = writer.Write(data)
	})
	log.Fatalln(http.ListenAndServe(":3359", nil))
}
