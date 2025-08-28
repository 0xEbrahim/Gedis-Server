package DB

import (
	"bufio"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Database struct {
	Aof  *os.File
	kv   map[string]string
	list map[string][]string
	hash map[string]map[string]string
	exp  map[string]time.Time
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
	db.mtx.Lock()
	if len(tokens) < 3 {
		return "-ERR: SET command required a key and a value\r\n"
	}
	db.kv[tokens[1]] = tokens[2]
	db.mtx.Unlock()
	return "+OK\r\n"
}
func (db *Database) Lpush(token []string) string {
	return ""
}
func (db *Database) Hset(tokens []string) string {
	return ""
}

func (db *Database) FlushAll(tokens []string) string {
	db.mtx.Lock()
	db.kv = make(map[string]string)
	db.list = make(map[string][]string)
	db.hash = make(map[string]map[string]string)
	db.mtx.Unlock()
	return "+OK\r\n"
}

func (db *Database) Keys(tokens []string) string {
	db.mtx.Lock()
	var keys []string
	for k, _ := range db.kv {
		keys = append(keys, k)
	}
	for k, _ := range db.list {
		keys = append(keys, k)
	}
	for k, _ := range db.hash {
		keys = append(keys, k)
	}
	db.mtx.Unlock()
	str := "*" + strconv.Itoa(len(keys)) + "\r\n"
	for _, it := range keys {
		str = str + "$" + strconv.Itoa(len(it)) + "\r\n" + it + "\r\n"
	}
	return str
}

func (db *Database) Type(tokens []string) string {
	db.mtx.Lock()
	if len(tokens) < 2 {
		return "-ERR: TYPE commands requires a key\r\n"
	}
	key := tokens[1]
	_, inKV := db.kv[key]
	_, inList := db.list[key]
	_, inHash := db.hash[key]
	exists := inKV || inList || inHash
	if !exists {
		return "_\r\n"
	}
	var str reflect.Type
	if inKV {
		str = reflect.TypeOf(db.kv[key])
	} else if inList {
		str = reflect.TypeOf(db.list[key])
	} else {
		str = reflect.TypeOf(db.hash[key])
	}
	db.mtx.Unlock()
	return "$" + strconv.Itoa(len(str.Name())) + "\r\n" + str.Name() + "\r\n"
}

func (db *Database) Del(tokens []string) string {
	db.mtx.Lock()
	if len(tokens) < 2 {
		return "-ERR: DEL|UNLINK commands requires a key\r\n"
	}
	delete(db.kv, tokens[1])
	delete(db.list, tokens[1])
	delete(db.hash, tokens[1])
	db.mtx.Unlock()

	return "+OK\r\n"
}

func (db *Database) Expire(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	if len(tokens) < 3 {
		return "-ERR: EXPIRE command requires a key and a time in seconds\r\n"
	}

	key := tokens[1]
	_, inKV := db.kv[key]
	_, inList := db.list[key]
	_, inHash := db.hash[key]
	exists := inKV || inList || inHash
	if !exists {
		return "-ERR: Key does not exist\r\n"
	}

	seconds, err := strconv.Atoi(tokens[2])
	if err != nil {
		return "-ERR: time must be an integer\r\n"
	}

	exp := time.Now().Add(time.Duration(seconds) * time.Second)
	db.exp[key] = exp

	return "#true\r\n"
}

func (db *Database) Rename(tokens []string) string {
	db.mtx.Lock()

	if len(tokens) < 3 {
		return "-ERR: RENAME command requires the old key and the new key"
	}
	return ""
}

func (db *Database) Get(tokens []string) string {
	db.mtx.Lock()
	if len(tokens) < 2 {
		return "-ERR: Get commands requires a key\r\n"
	}
	str, ok := db.kv[tokens[1]]
	db.mtx.Unlock()
	var ret string
	if ok {
		ret = "$" + strconv.Itoa(len(str)) + "\r\n" + str + "\r\n"
	} else {
		ret = "_\r\n"
	}
	return ret
}
