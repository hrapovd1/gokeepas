package crypto

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/hrapovd1/gokeepas/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenX509KeyPair(t *testing.T) {
	crt, err := GenX509KeyPair()
	require.NoError(t, err)
	assert.Equal(t, 1, len(crt.Certificate))
}

func TestGenSymmKey(t *testing.T) {
	keyLen := 16
	out, err := GenSymmKey(keyLen)
	require.NoError(t, err)
	assert.Equal(t, len(out), keyLen)
}

func TestGenServerKey(t *testing.T) {
	keyLen := 16
	key, err := GenServerKey(keyLen)
	require.NoError(t, err)
	assert.Equal(t, keyLen, len(key))
}

func TestEncryptKey(t *testing.T) {
	symmKey := []byte(`qwcsposfJOshf.34jswo_sdf`)
	data := []byte("12345")
	result, err := EncryptKey(symmKey, data)
	require.NoError(t, err)
	encData, err := base64.StdEncoding.DecodeString(result)
	require.NoError(t, err)
	assert.Equal(t, 33, len(encData))
}

func TestDecryptKey(t *testing.T) {
	encStr := "aV6TS1ylt+Y0UrlimwY0lwqdZeZh1w5f1+wFOvY4eZPv"
	symmKey := []byte(`qwcsposfJOshf.34jswo_sdf`)
	result, err := DecryptKey(symmKey, encStr)
	require.NoError(t, err)
	assert.Equal(t, []byte("12345"), result)
}

func TestHashPasswd(t *testing.T) {
	passwd := []byte("sdfwerJ.45fj")
	passwdHash := "2cec73172dedd21e866ce3ec51011065d36656fc"
	result, err := HashPasswd(context.Background(), passwd)
	require.NoError(t, err)
	assert.Equal(t, passwdHash, result)
}

func TestGetToken(t *testing.T) {
	login := "test"
	passwd := "sdfwerJ.45fj"
	userData := types.StorageModel{
		PassHash: "2cec73172dedd21e866ce3ec51011065d36656fc",
	}
	key := []byte("12345Tre.wq")
	t.Run("right", func(t *testing.T) {
		token, err := GetToken(context.Background(), login, passwd, userData, key)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

	})
	t.Run("wrong", func(t *testing.T) {
		token, err := GetToken(context.Background(), login, "", userData, key)
		require.Error(t, err)
		assert.Empty(t, token)
	})
}

func TestCheckToken(t *testing.T) {
	oldToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6InRlc3QiLCJleHAiOjE2ODQyNDEyMTcsImlhdCI6MTY4NDIzOTQxN30.DN_yTLLQqpnpacauLXYAZ23lNq-iA3mkPsPa5np6shM"
	login := "test"
	passwd := "sdfwerJ.45fj"
	userData := types.StorageModel{
		PassHash: "2cec73172dedd21e866ce3ec51011065d36656fc",
	}
	key := []byte("12345Tre.wq")
	token, err := GetToken(context.Background(), login, passwd, userData, key)
	require.NoError(t, err)
	t.Run("right", func(t *testing.T) {
		result, err := CheckToken(token, key)
		require.NoError(t, err)
		assert.Equal(t, login, result)
	})
	t.Run("old token", func(t *testing.T) {
		result, err := CheckToken(oldToken, key)
		require.Error(t, err)
		assert.Empty(t, result)
	})
	t.Run("wrong token", func(t *testing.T) {
		lenToken := len(oldToken)
		wrongToken := oldToken[:(lenToken-1)/2-4] + oldToken[(lenToken-1)/2:]
		result, err := CheckToken(wrongToken, key)
		require.Error(t, err)
		assert.Empty(t, result)
	})

}
