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
	kv   map[string]string
	list map[string][]string
	hash map[string]map[string]string
	mtx  *sync.Mutex
}

var instance *Database
var once sync.Once

func GetDBInstance() *Database {
	once.Do(func() {
		file, err := os.OpenFile("db.log", os.O_RDWR|os.O_CREATE, 0644)
		if err != nil || file == nil {
			log.Fatal("Unable to open AOF")
		}
		instance = &Database{Aof: file, kv: map[string]string{}, list: map[string][]string{}, hash: map[string]map[string]string{}}

	})
	return instance
}

func (db *Database) Flush() {
	db.mtx.Lock()
	println("Flush")
	err := db.Aof.Truncate(0)
	if err != nil {
		println("Error trunc")
		return
	}
	db.Aof.Seek(0, 0)
	defer db.Aof.Close()
	for k, v := range db.kv {
		str := "SET " + k + " " + (v) + "\n"
		_, err := db.Aof.Write([]byte(str))
		if err != nil {
			println("Error while writing AOF", err.Error())
			return
		}
	}
	for k, v := range db.list {
		_, err := db.Aof.Write([]byte("LPUSH"))
		if err != nil {
			println("Error while writing AOF", err.Error())
			return
		}
		_, err = db.Aof.Write([]byte(" " + k))
		if err != nil {
			println("Error while writing AOF", err.Error())
			return
		}
		for _, it := range v {
			_, err = db.Aof.Write([]byte(" " + it))
			if err != nil {
				println("Error while writing AOF", err.Error())
				return
			}
		}
		_, err = db.Aof.Write([]byte("\n"))
		if err != nil {
			println("Error while writing AOF", err.Error())
			return
		}
	}
	for k, v := range db.hash {
		for K, V := range v {
			str := "HSET " + k + " " + K + " " + V + "\n"
			_, err := db.Aof.Write([]byte(str))
			if err != nil {
				println("Error while writing AOF", err.Error())
				return
			}
		}
	}
	db.mtx.Unlock()
}
func (db *Database) Load() {
	db.mtx.Lock()
	println("LOAD")
	sc := bufio.NewScanner(db.Aof)
	for sc.Scan() {
		line := sc.Text()
		tokens := strings.Split(line, " ")
		switch tokens[0] {
		case "SET":
			db.Set(tokens)
		case "LPUSH":
			db.Lpush(tokens)
		case "HSET":
			db.Hset(tokens)
		default:

		}
	}
	db.mtx.Unlock()
}

func (db *Database) Set(tokens []string) string {
	db.kv[tokens[1]] = tokens[2]
	return "+OK\r\n"
}
func (db *Database) Lpush(token []string) {

}
func (db *Database) Hset(tokens []string) {}
func (db *Database) Get(tokens []string) string {
	str := db.kv[tokens[1]]
	ret := "$" + strconv.Itoa(len(str)) + "\r\n" + str + "\r\n"
	return ret
}
