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
		instance = &Database{Aof: file, kv: map[string]string{}, list: map[string][]string{}, hash: map[string]map[string]string{}, exp: map[string]time.Time{}, mtx: &sync.Mutex{}}

	})
	return instance
}

func (db *Database) Flush() {
	db.mtx.Lock()
	defer db.mtx.Unlock()
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
		_, err := db.Aof.Write([]byte("RPUSH"))
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

}
func (db *Database) Load() {
	println("LOAD")
	sc := bufio.NewScanner(db.Aof)
	for sc.Scan() {
		line := sc.Text()
		tokens := strings.Split(line, " ")
		switch tokens[0] {
		case "SET":
			db.Set(tokens)
		case "RPUSH":
			db.RPush(tokens)
		case "HSET":
			db.HSet(tokens)
		default:

		}
	}

}

func (db *Database) Set(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	if len(tokens) < 3 {
		return "-ERR: SET command required a key and a value\r\n"
	}
	db.kv[tokens[1]] = tokens[2]
	return "+OK\r\n"
}

func (db *Database) FlushAll(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.kv = make(map[string]string)
	db.list = make(map[string][]string)
	db.hash = make(map[string]map[string]string)

	return "+OK\r\n"
}

func (db *Database) Keys(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()

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
	str := "*" + strconv.Itoa(len(keys)) + "\r\n"
	for _, it := range keys {
		str = str + "$" + strconv.Itoa(len(it)) + "\r\n" + it + "\r\n"
	}
	return str
}

func (db *Database) Type(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()
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
	return "$" + strconv.Itoa(len(str.Name())) + "\r\n" + str.Name() + "\r\n"
}

func (db *Database) Del(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	if len(tokens) < 2 {
		return "-ERR: DEL|UNLINK commands requires a key\r\n"
	}
	delete(db.kv, tokens[1])
	delete(db.list, tokens[1])
	delete(db.hash, tokens[1])

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
	defer db.mtx.Unlock()
	if len(tokens) < 3 {
		return "-ERR: RENAME command requires the old key and the new key"
	}
	key := tokens[1]
	_, inKV := db.kv[key]
	_, inList := db.list[key]
	_, inHash := db.hash[key]
	exists := inKV || inList || inHash
	if !exists {
		return "-ERR: Key does not exist\r\n"
	}
	if inKV {
		value := db.kv[key]
		delete(db.kv, key)
		db.kv[tokens[2]] = value
	} else if inList {
		value := db.list[key]
		delete(db.list, key)
		db.list[tokens[2]] = value
	} else {
		value := db.hash[key]
		delete(db.hash, key)
		db.hash[tokens[2]] = value
	}
	return "#true\r\n"
}

func (db *Database) Get(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	if len(tokens) < 2 {
		return "-ERR: Get commands requires a key\r\n"
	}
	str, ok := db.kv[tokens[1]]
	var ret string
	if ok {
		ret = "$" + strconv.Itoa(len(str)) + "\r\n" + str + "\r\n"
	} else {
		ret = "_\r\n"
	}
	return ret
}

func (db *Database) LLen(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	if len(tokens) < 2 {
		return "-ERR: LLEN command requires a key\r\n"
	}
	key := tokens[1]
	v, ok := db.list[key]
	if !ok {
		return "_\r\n"
	}
	return ":" + strconv.Itoa(len(v)) + "\r\n"
}

func (db *Database) LPush(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	if len(tokens) < 3 {
		return "-ERR: LPUSH command requires a key and a value\r\n"
	}

	v, _ := db.list[tokens[1]]
	for i := 2; i < len(tokens); i++ {

		v = append(
			[]string{
				tokens[i],
			}, v...)
	}
	db.list[tokens[1]] = v
	return ":" + strconv.Itoa(len(v)) + "\r\n"
}

func (db *Database) RPush(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	if len(tokens) < 3 {
		return "-ERR: RPUSH command requires a key and a value\r\n"
	}
	v, _ := db.list[tokens[1]]

	for i := 2; i < len(tokens); i++ {
		v = append(
			v, tokens[i])
	}
	db.list[tokens[1]] = v
	return ":" + strconv.Itoa(len(v)) + "\r\n"

}
func (db *Database) LPop(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	if len(tokens) < 2 {
		return "-ERR: LPOP command requires a key\r\n"
	}
	v, ok := db.list[tokens[1]]
	if !ok || len(v) == 0 {
		return "_\r\n"
	}
	value := v[0]
	v = v[1:]
	db.list[tokens[1]] = v
	return "$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n"
}
func (db *Database) RPop(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	if len(tokens) < 2 {
		return "-ERR: RPOP command requires a key\r\n"
	}
	v, ok := db.list[tokens[1]]
	if !ok || len(v) == 0 {
		return "_\r\n"
	}
	value := v[len(v)-1]
	v = v[:len(v)-1]
	db.list[tokens[1]] = v
	return "$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n"
}
func (db *Database) LRem(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	if len(tokens) < 4 {
		return "-ERR: LREM command requires key, count and a value\r\n"
	}
	key := tokens[1]
	count, err := strconv.Atoi(tokens[2])
	if err != nil {
		return "-ERR: count must be a number\r\n"
	}
	value := tokens[3]
	removed := 0
	v, ok := db.list[tokens[1]]
	if !ok {
		return ":0\r\n"
	}
	if count == 0 {
		var nList []string
		for i := 0; i < len(v); i++ {
			if v[i] != value {
				nList = append(nList, v[i])
			} else {
				removed++
			}
		}
		db.list[key] = nList
		return ":" + strconv.Itoa(removed) + "\r\n"
	} else if count > 0 {
		var nList []string
		for i := 0; i < len(v); i++ {
			if v[i] != value {
				nList = append(nList, v[i])
			} else {
				if removed == count {
					nList = append(nList, v[i])
				}
				removed = removed + 1
			}
		}
		db.list[key] = nList
		return ":" + strconv.Itoa(removed) + "\r\n"
	} else {
		var nList []string
		for i := len(v) - 1; i >= 0; i-- {
			if v[i] != value {
				nList = append(nList, v[i])
			} else {
				if removed == count {
					nList = append(nList, v[i])
				}
				removed = removed - 1
			}
		}
		db.list[key] = nList
		return ":" + strconv.Itoa(removed) + "\r\n"
	}
}
func (db *Database) LIndex(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	if len(tokens) < 3 {
		return "-ERR: LINDEX command requires a key and an index\r\n"
	}
	key := tokens[1]
	index, err := strconv.Atoi(tokens[2])
	if err != nil {
		return "-ERR: index must be a number\r\n"
	}
	v, ok := db.list[key]
	if !ok {
		return ":0\r\n"
	}
	n := len(v)
	if index < 0 {
		index = index + n
	}
	if index >= n || index < 0 {
		return "_\r\n"
	}
	value := v[index]
	return "$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n"
}

func (db *Database) LSet(tokens []string) string {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	if len(tokens) < 4 {
		return "-ERR: LSET requires key, index and a value"
	}
	key := tokens[1]
	index, err := strconv.Atoi(tokens[2])
	if err != nil {
		return "-ERR: index must be a number\r\n"
	}
	v, ok := db.list[key]
	if !ok {
		return ":0\r\n"
	}
	n := len(v)
	if index < 0 {
		index = index + n
	}
	if index >= n || index < 0 {
		return "_\r\n"
	}
	db.list[key][index] = tokens[3]
	return "+OK\r\n"
}
func encodeArray(tokens []string) string {
	encoded := "*" + strconv.Itoa(len(tokens)) + "\r\n"
	for _, it := range tokens {
		encoded = encoded + "$" + strconv.Itoa(len(it)) + "\r\n" + it + "\r\n"
	}
	return encoded
}

func (db *Database) LRange(tokens []string) string {
	if len(tokens) < 4 {
		return "-ERR: LRANGE command requires key, start and an end\r\n"
	}
	key := tokens[1]
	lst, ok := db.list[key]
	if !ok {
		return "_\r\n"
	}
	start, err := strconv.Atoi(tokens[2])
	if err != nil {
		return "-ERR: start must be an integer\r\n"
	}
	end, err := strconv.Atoi(tokens[3])
	if err != nil {
		return "-ERR: end must be an integer\r\n"
	}
	if start < 0 {
		start = start + len(lst)
	}
	if end < 0 {
		end = end + len(lst)
	}
	if start > end {
		return "_\r\n"
	}
	end = min(end, len(lst)-1)
	start = max(0, start)
	var arr []string
	for i := start; i <= end; i++ {
		arr = append(arr, lst[i])
	}
	return encodeArray(arr)
}

func (db *Database) HSet(tokens []string) string {
	if len(tokens) < 4 || len(tokens)%2 == 1 {
		return "-ERR: HSET command requires key, field and a value\r\n"
	}

	return ":1\r\n"
}
func (db *Database) HGet(tokens []string) string {
	if len(tokens) < 3 {
		return "-ERR: HGET command requires key and field \r\n"
	}
	return ""
}

func (db *Database) HExists(tokens []string) string {
	if len(tokens) < 3 {
		return "-ERR: HEXISTS command requires a key and a field\r\n"
	}
	return ":1\r\n"
}

func (db *Database) HDel(tokens []string) string {
	if len(tokens) < 3 {
		return "-ERR: HDEL command requires a key and at least one field\r\n"
	}
	return ":1\r\n"
}

func (db *Database) HGetAll(tokens []string) string {
	if len(tokens) < 2 {
		return "-ERR: HGETALL command requires a key\r\n"
	}
	return ""
}

func (db *Database) HKeys(tokens []string) string {
	if len(tokens) < 2 {
		return "-ERR: HKEYS command requires a key\r\n"
	}
	return ""

}

func (db *Database) HVals(tokens []string) string {
	if len(tokens) < 2 {
		return "-ERR: HVALS command requires a key\r\n"
	}
	return ""

}
func (db *Database) HLen(tokens []string) string {
	if len(tokens) < 2 {
		return "-ERR: HLEN command requires a key\r\n"
	}
	return ""
}

func (db *Database) HMSet(tokens []string) string {
	if len(tokens) < 4 || len(tokens)%2 == 1 {
		return "-ERR: HMSET command requires key, field and a value\r\n"
	}

	return ":OK\r\n"
}
