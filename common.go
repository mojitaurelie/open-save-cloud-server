package main

import (
	"io"
	"log"
	"opensavecloudserver/config"
	"opensavecloudserver/database"
	"os"
)

func InitCommon() {
	f, err := os.OpenFile("server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println(err)
		}
	}(f)
	log.SetOutput(io.MultiWriter(os.Stdout, f))

	config.Init()
	database.Init()
}
