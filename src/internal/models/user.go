package models

type User struct {
	CommonFields

	Username     string `gorm:"unique"`
	PasswordHash string
	Is_Super     bool
	Email        string
	DisplayName  string
	AvatarURL    string
	Bio          string
}
