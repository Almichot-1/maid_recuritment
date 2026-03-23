package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
)

func main() {
	file := flag.String("file", "", "SQL migration file to execute")
	flag.Parse()

	if *file == "" {
		log.Fatal("missing -file")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	content, err := os.ReadFile(*file)
	if err != nil {
		log.Fatalf("read migration file: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}

	if err := db.Exec(string(content)).Error; err != nil {
		log.Fatalf("apply migration: %v", err)
	}

	fmt.Printf("Applied migration: %s\n", *file)
}
