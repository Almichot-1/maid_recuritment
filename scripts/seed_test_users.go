package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID           string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	FullName     string
	Role         string `gorm:"type:user_role;not null"`
	CompanyName  string
	IsActive     bool `gorm:"default:true"`
}

func main() {
	// Connect to local Docker PostgreSQL
	dsn := "postgresql://postgres:postgres@127.0.0.1:5432/maid_recruitment?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	password := "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Ethiopian Agent
	ethiopian := User{
		ID:           "11111111-1111-1111-1111-111111111111",
		Email:        "ethiopian@test.com",
		PasswordHash: string(hash),
		FullName:     "Ethiopian Test Agent",
		Role:         "ethiopian_agent",
		CompanyName:  "Ethiopian Recruitment Agency",
		IsActive:     true,
	}

	// Foreign Agent
	foreign := User{
		ID:           "22222222-2222-2222-2222-222222222222",
		Email:        "foreign@test.com",
		PasswordHash: string(hash),
		FullName:     "Foreign Test Agent",
		Role:         "foreign_agent",
		CompanyName:  "International Recruitment Co",
		IsActive:     true,
	}

	// Insert users
	if err := db.Table("users").Create(&ethiopian).Error; err != nil {
		log.Printf("Ethiopian agent may already exist: %v", err)
	} else {
		fmt.Println("✓ Created Ethiopian Agent")
	}

	if err := db.Table("users").Create(&foreign).Error; err != nil {
		log.Printf("Foreign agent may already exist: %v", err)
	} else {
		fmt.Println("✓ Created Foreign Agent")
	}

	fmt.Println("\n=== Test Credentials ===")
	fmt.Println("\nEthiopian Agent (Creates Candidates):")
	fmt.Println("  Email: ethiopian@test.com")
	fmt.Println("  Password: password123")
	fmt.Println("  Company: Ethiopian Recruitment Agency")
	fmt.Println("\nForeign Agent (Selects Candidates):")
	fmt.Println("  Email: foreign@test.com")
	fmt.Println("  Password: password123")
	fmt.Println("  Company: International Recruitment Co")
	fmt.Println("\n========================")
}
