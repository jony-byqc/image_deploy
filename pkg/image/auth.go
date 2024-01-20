package image

import (
	"encoding/base64"
	"strings"
)

type Auth interface {
	ParseAuthHeader() string
}

type BasicAuth struct {
	UserName string
	PassWord string
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (b *BasicAuth) ParseAuthHeader() string {
	return "Basic " + basicAuth(b.UserName, b.PassWord)
}

func DecodeBasicAuth(authHex string) Auth {
	decodeBytes, err := base64.StdEncoding.DecodeString(authHex)
	if err != nil {
		return nil
	}

	args := strings.Split(string(decodeBytes), ":")
	if len(args) > 1 {
		return &BasicAuth{
			UserName: args[0],
			PassWord: args[1],
		}
	}

	return nil
}
