package model

import (
	"time"
)

type SimpleAccountData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type InitialData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DataType    string `json:"data_type"`
	Data        string `json:"data"`
}

type Metadata struct {
	Name        string    `json:"name"`
	Created     time.Time `json:"created"`
	Changed     time.Time `json:"changed"`
	Description string    `json:"description"`
	StaticID    string    `json:"static_id"`
	DynamicID   string    `json:"dynamic_id"`
	UserID      string    `json:"user_id"`
	DataType    string    `json:"data_type"`
}

type LoginAndPasswordData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type CardData struct {
	CardNumber     string `json:"card_number"`
	ExpiredAt      string `json:"expired_at"`
	CardholderName string `json:"cardholder_name"`
	Code           string `json:"code"`
}

type SimpleMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DataType    string `json:"data_type"`
}

type DataToDelete struct {
	StaticID string `json:"static_id"`
	UserID   string `json:"user_id"`
	DataType string `json:"data_type"`
}

type DataToRead struct {
	StaticID string `json:"static_id"`
	UserID   string `json:"user_id"`
	DataType string `json:"data_type"`
}

type ReadResponse struct {
	StaticID string `json:"static_id"`
	Data     string `json:"data"`
}

type EditData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DataType    string `json:"data_type"`
	Data        string `json:"data"`
	StaticID    string `json:"static_id"`
	UserID      string `json:"user_id"`
}
