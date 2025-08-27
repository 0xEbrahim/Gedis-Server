package AOF

import (
	"log"
	"os"
)

type AOF struct {
	Aof *os.File
}

func InitAOF() *AOF {
	file, err := os.OpenFile("aof.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal("Unable to open AOF")
	}
	return &AOF{Aof: file}
}

func (aof *AOF) Flush() {

}

func (aof *AOF) Load() {

}
