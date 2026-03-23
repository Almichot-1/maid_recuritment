package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/internal/service"
)

type seedOutput struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	Role            string `json:"role"`
	MFASecret       string `json:"mfa_secret"`
	ProvisioningURL string `json:"provisioning_url"`
	CurrentOTP      string `json:"current_otp"`
}

func main() {
	email := flag.String("email", "admin@test.com", "admin email")
	password := flag.String("password", "AdminPassword123!", "admin password")
	fullName := flag.String("name", "Platform Super Admin", "admin full name")
	role := flag.String("role", "super_admin", "admin role")
	mfaSecret := flag.String("mfa-secret", "", "optional MFA secret")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	adminRepository, err := repository.NewGormAdminRepository(cfg)
	if err != nil {
		log.Fatalf("init admin repository: %v", err)
	}
	auditRepository, err := repository.NewGormAuditLogRepository(cfg)
	if err != nil {
		log.Fatalf("init audit repository: %v", err)
	}
	authService, err := service.NewAdminAuthService(adminRepository, auditRepository, cfg)
	if err != nil {
		log.Fatalf("init admin auth service: %v", err)
	}

	secret := *mfaSecret
	provisioningURL := ""
	if secret == "" {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "Maid Recruitment Platform",
			AccountName: *email,
			Algorithm:   otp.AlgorithmSHA1,
			Digits:      otp.DigitsSix,
		})
		if err != nil {
			log.Fatalf("generate mfa secret: %v", err)
		}
		secret = key.Secret()
		provisioningURL = key.URL()
	}

	admin, err := authService.CreateAdmin(*email, *password, *fullName, *role, secret)
	if err != nil {
		if !errors.Is(err, service.ErrUserExists) {
			log.Fatalf("create admin: %v", err)
		}

		admin, err = adminRepository.GetByEmail(*email)
		if err != nil {
			log.Fatalf("load existing admin: %v", err)
		}

		hash, hashErr := bcrypt.GenerateFromPassword([]byte(*password), 12)
		if hashErr != nil {
			log.Fatalf("hash password: %v", hashErr)
		}

		admin.Email = *email
		admin.FullName = *fullName
		admin.PasswordHash = string(hash)
		admin.MFASecret = secret
		admin.IsActive = true
		admin.Role = domain.AdminRole(*role)
		admin.LockedUntil = nil
		admin.FailedLoginAttempts = 0
		admin.ForcePasswordChange = false

		if err := adminRepository.Update(admin); err != nil {
			log.Fatalf("update existing admin: %v", err)
		}
	}

	admin.ForcePasswordChange = false
	if err := adminRepository.Update(admin); err != nil {
		log.Fatalf("finalize admin: %v", err)
	}

	currentOTP, err := totp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		log.Fatalf("generate current otp: %v", err)
	}

	if provisioningURL == "" {
		key, err := otp.NewKeyFromURL("otpauth://totp/Maid%20Recruitment%20Platform:" + *email + "?secret=" + secret + "&issuer=Maid%20Recruitment%20Platform")
		if err == nil {
			provisioningURL = key.URL()
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(seedOutput{
		Email:           admin.Email,
		Password:        *password,
		Role:            string(admin.Role),
		MFASecret:       secret,
		ProvisioningURL: provisioningURL,
		CurrentOTP:      currentOTP,
	}); err != nil {
		log.Fatalf("write output: %v", err)
	}
}
