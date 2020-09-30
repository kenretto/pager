# pager ![build](https://api.travis-ci.org/kenretto/pager.svg?branch=master&status=passed)
Pager is a web paging tool for golang

# Interface 
This tool provides gorm-based data paging. If you don't need Gorm, you can choose to implement the following interface or create issues.

```go
type Driver interface {
	Where(kv Where)
	Range(r Range)
	Limit(limit int)
	Skip(skip int)
	Index(index string)
	Sort(kv []SortInfo)
	Find(data interface{})
	SetTyp(typ reflect.Type)
	Count() int64
}
```

# Url-rule
- http://localhost:3359/member?range=id:14 
- http://localhost:3359/member?name=a
- http://localhost:3359/member?sort:-age
- http://localhost:3359/member?rows=3
- http://localhost:3359/member?page=2
- http://localhost:3359/member?range=id:14,23&rows=1&page=2&sort:-age
- http://localhost:3359/member?range=id:3,8&range=age:6,4

Run the sample code to see the effect 

# example
```go
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

```