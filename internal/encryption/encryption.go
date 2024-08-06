package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v4"
)

const (
	dataKeyLength = 16
	fileWithKey   = "internal/sk/encryption.txt"
)

type Claims struct {
	jwt.RegisteredClaims
	Data string
}

// эти методы перенести в другой пакет
func GenerateSK(data string) (string, error) {
	var newKey string

	if len(data) <= 0 {
		return "", errors.New("expected length is not valid")
	}

	hash := make([]byte, dataKeyLength)
	_, err := rand.Read(hash)
	if err != nil {
		return "", err
	}

	encodedHash := base64.StdEncoding.EncodeToString(hash)
	newKey = encodedHash[:dataKeyLength]

	encryptionSK, err := getEncryptionKeyFromFile(fileWithKey)
	if err != nil {
		return "", err
	}
	encryptedDataSK, err := buildJWTString(encryptionSK, newKey)
	if err != nil {
		return "", err
	}

	return encryptedDataSK, nil
}

func EncryptSimpleData(sk string, data string) (string, error) {
	return buildJWTString(sk, data)
}

func DecryptData(dataSK string, encryptedData string) (string, error) {
	realDataSk, err := decryptDataSK(dataSK)
	
	if err != nil {
		return "", err
	}
	claims := new(Claims)
	token, err := jwt.ParseWithClaims(encryptedData, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(realDataSk), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		log.Printf("Token is not valid")
		return "", errors.New("token is not valid")
	}

	return claims.Data, nil
}

func decryptDataSK(dataSK string) (string, error) {
	encryptionKey, err := getEncryptionKeyFromFile(fileWithKey)
	if err != nil {
		return "", err
	}
	claims := new(Claims)
	token, err := jwt.ParseWithClaims(dataSK, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(encryptionKey), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		log.Printf("Token is not valid")
		return "", errors.New("token is not valid")
	}

	return claims.Data, nil
}

func buildJWTString(sk string, data string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		// собственное утверждение
		Data: data,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(sk))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func getEncryptionKeyFromFile(filename string) (string, error) {
	var encryptionKey string
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	encryptionKey = string(content)

	return encryptionKey, nil
}
