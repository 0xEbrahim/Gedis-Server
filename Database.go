package main

import (
	"Gedis-Server/AOF"
	"sync"
)

type Database struct {
	aof *AOF.AOF
}

var instance *Database
var once sync.Once

func getDBInstance() *Database {
	once.Do(func() {
		instance = &Database{aof: AOF.InitAOF()}
	})
	return instance
}

func (db *Database) flush() {
	db.aof.Flush()
}
func (db *Database) load() {
	db.aof.Load()
}
