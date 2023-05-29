/*
Package crypto contents type and methods for encryption/decryption
*/
package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hrapovd1/gokeepas/internal/types"
)

const (
	hashSalt       = "oerwtOUFHsa.sd!df56s"
	expireDuration = 30 * time.Minute // jwt token live time

	alphabet      = 61
	SymmKeyLength = 24 // length of user symmetric key
)

// GenX509KeyPair generates the TLS keypair for the server
// https://gist.github.com/shivakar/cd52b5594d4912fbeb46
func GenX509KeyPair() (tls.Certificate, error) {
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(now.Unix()),
		Subject: pkix.Name{
			CommonName:         "gokeepas.local.net",
			Country:            []string{"RU"},
			Organization:       []string{"local.net"},
			OrganizationalUnit: []string{"gokeepas"},
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(1, 0, 0), // Valid for one year
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		DNSNames: []string{"*"},
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, template,
		priv.Public(), priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	var outCert tls.Certificate
	outCert.Certificate = append(outCert.Certificate, cert)
	outCert.PrivateKey = priv

	return outCert, nil
}

// GenKey return symmetric key for user when signup and for server
func GenSymmKey(n int) ([]byte, error) {
	out := make([]byte, n)
	n1, err := rand.Read(out)
	if err != nil || n1 != n {
		return out, err
	}
	return out, nil
}

// GenServerKey generates master server key when it isn't provided
func GenServerKey(n int) (string, error) {
	dict := []rune("A7bBaCd8DeE3fFjGi2HkIlJmK9oLpMz4Nx0OyPwQgRnS5qTqrUs6VtWuXvYZ1")
	out := make([]rune, 0)
	tmp := make([]byte, 1)
	for i := 0; i < n; i++ {
		if _, err := rand.Read(tmp); err != nil {
			return "", err
		}
		out = append(out, dict[int(tmp[0])%alphabet])
	}
	return string(out), nil
}

// EncryptKey encrypt data with symmKey
func EncryptKey(symmKey []byte, data []byte) (string, error) {
	// encrypt data
	cphr, err := aes.NewCipher(symmKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(cphr)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	dataEnc := gcm.Seal(nonce, nonce, data, nil)

	return base64.StdEncoding.EncodeToString(dataEnc), nil
}

// DecryptKey decrypt data with symmKey
func DecryptKey(symm []byte, data string) ([]byte, error) {
	encJSON, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	// Decrypt symm key from db
	chpr, err := aes.NewCipher(symm)
	if err != nil {
		return nil, err
	}
	gcmDecrypt, err := cipher.NewGCM(chpr)
	if err != nil {
		return nil, err
	}
	nonceSize := gcmDecrypt.NonceSize()
	if len(encJSON) < nonceSize {
		return nil, errors.New("len(encJSON) < nonceSize")
	}
	nonce, encDataJSON := encJSON[:nonceSize], encJSON[nonceSize:]
	return gcmDecrypt.Open(nil, nonce, encDataJSON, nil)
}

// HashPasswd return hash of password
func HashPasswd(_ context.Context, passwd []byte) (string, error) {
	pwdHash := sha1.New()
	pwdHash.Write(passwd)
	pwdHash.Write([]byte(hashSalt))

	return fmt.Sprintf("%x", pwdHash.Sum(nil)), nil
}

// GetToken generate jwt session token for user
func GetToken(_ context.Context, login string, passwd string, userData types.StorageModel, key []byte) (string, error) {
	pwdHash := sha1.New()
	pwdHash.Write([]byte(passwd))
	pwdHash.Write([]byte(hashSalt))
	password := fmt.Sprintf("%x", pwdHash.Sum(nil))
	if userData.PassHash != password {
		return "", fmt.Errorf("wrong login or password")
	}
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&types.Claims{
			Login: login,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		})

	return token.SignedString(key)
}

// CheckToken checks jwt token is provided by cli
func CheckToken(tkn string, key []byte) (string, error) {
	token, err := jwt.ParseWithClaims(
		tkn,
		&types.Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return key, nil
		},
	)
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*types.Claims); ok && token.Valid {
		return claims.Login, nil
	}

	return "", errors.New("token wrong")
}
