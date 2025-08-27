package DB

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Database struct {
	Aof  *os.File
	data map[string]string
}

var instance *Database
var once sync.Once

func GetDBInstance() *Database {
	once.Do(func() {
		file, err := os.OpenFile("db.log", os.O_RDWR|os.O_CREATE, 0644)
		if err != nil || file == nil {
			log.Fatal("Unable to open AOF")
		}
		instance = &Database{Aof: file, data: map[string]string{}}

	})
	return instance
}

func (db *Database) Flush() {
	println("Flush")
	err := db.Aof.Truncate(0)
	if err != nil {
		println("Error trunc")
		return
	}
	db.Aof.Seek(0, 0)
	data := db.data
	defer db.Aof.Close()
	for k, v := range data {
		str := "SET " + k + " " + (v) + "\n"

		_, err := db.Aof.Write([]byte(str))

		if err != nil {
			println("Error while writing AOF", err.Error())
			return
		}
	}
}
func (db *Database) Load() {
	println("LOAD")
	sc := bufio.NewScanner(db.Aof)
	for sc.Scan() {
		line := sc.Text()
		tokens := strings.Split(line, " ")
		db.Set(tokens)
	}
}

func (db *Database) Set(tokens []string) string {
	db.data[tokens[1]] = tokens[2]
	return "+OK\r\n"
}
func (db *Database) Get(tokens []string) string {
	str := db.data[tokens[1]]
	ret := "$" + strconv.Itoa(len(str)) + "\r\n" + str + "\r\n"
	return ret
}
