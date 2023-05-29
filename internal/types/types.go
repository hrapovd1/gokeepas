/*
Package types contents types for server and client.
*/
package types

import "github.com/golang-jwt/jwt/v4"

// Claims type for jwt tokens
type Claims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

// Login type implements login secret.
type Login struct {
	Login    string   `json:"login"`
	Password string   `json:"password,omitempty"`
	Info     []string `json:"info,omitempty"`
}

// Text implements text secret.
type Text struct {
	Text string   `json:"text"`
	Info []string `json:"info,omitempty"`
}

// Binary implements binary secret.
type Binary struct {
	Data []byte   `json:"data"`
	Info []string `json:"info,omitempty"`
}

// Cart implements bank plastic cart secret.
type Cart struct {
	Number  string   `json:"number"`
	Expired string   `json:"expired"`
	Holder  string   `json:"holder"`
	CVC     string   `json:"cvc"`
	Info    []string `json:"info,omitempty"`
}

// StorageModel implements storage db model.
type StorageModel struct {
	PassHash string `redis:"pass"`
	SymmKey  string `redis:"symmkey"`
	Data     string `redis:"data"`
	Type     string `redis:"type"`
}
