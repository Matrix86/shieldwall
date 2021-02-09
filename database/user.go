package database

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/badoux/checkmail"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/str"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"time"
)

const MinPasswordLength = 8

// TODO: add 2FA
type User struct {
	ID           uint           `gorm:"primarykey"`
	CreatedAt    time.Time      `gorm:"index"`
	UpdatedAt    time.Time      `gorm:"index"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Email        string         `gorm:"index"`
	Verification string         `gorm:"index"`
	Verified     bool           `gorm:"index"`
	Hash         string
	Address      string
	Agents       []Agent
}

func makeVerification() string {
	randomShit := make([]byte, 128)
	rand.Read(randomShit)

	data := append(
		[]byte(strconv.FormatInt(time.Now().UnixNano(), 10)),
		randomShit...)

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func RegisterUser(address, email, password string) (*User, error) {
	if err := checkmail.ValidateFormat(email); err != nil {
		return nil, err
	} else if password = str.Trim(password); len(password) < MinPasswordLength {
		return nil, fmt.Errorf("minimum password length is %d", MinPasswordLength)
	}

	var found User
	if err := db.Where("email=?", email).First(&found).Error; err == nil {
		return nil, fmt.Errorf("email address already used")
	} else if err != gorm.ErrRecordNotFound {
		log.Error("error searching email '%s': %v", email, err)
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error generating password hash: %v", err)
	}

	newUser := User{
		Email:        email,
		Verification: makeVerification(),
		Hash:         string(hashedPassword),
		Address:      address,
	}

	if err = db.Create(&newUser).Error; err != nil {
		return nil, fmt.Errorf("error creating new user: %v", err)
	}

	return &newUser, nil
}

func VerifyUser(verification string) error {
	var found User
	if err := db.Where("verification=?", verification).First(&found).Error; err != nil {
		return err
	} else if found.Verified == true {
		return fmt.Errorf("user already verified")
	} else {
		found.Verified = true
		return db.Save(&found).Error
	}
}

func LoginUser(address, email, password string) (*User, error) {
	var found User
	if err := db.Where("email=?", email).First(&found).Error; err == gorm.ErrRecordNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else if found.Verified == false {
		return nil, fmt.Errorf("account not verified")
	} else if err = bcrypt.CompareHashAndPassword([]byte(found.Hash), []byte(password)); err != nil {
		return nil, nil
	}

	found.Address = address
	found.UpdatedAt = time.Now()
	if err := db.Save(&found).Error; err != nil {
		log.Error("error updating logged in user: %v", err)
	}

	return &found, nil
}