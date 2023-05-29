package cli

import (
	"testing"

	"github.com/hrapovd1/gokeepas/internal/config"
	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_runGet(t *testing.T) {
	client := cliClient{logger: zap.New(nil), config: config.Config{}}
	t.Run("empty args", func(t *testing.T) {
		runGet(&client, false, &cobra.Command{}, []string{})
	})
}

func Test_printText(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		jsonFmt  bool
		positive bool
	}{
		{"textOut", `{"text":"one", "info":["extra"]}`, false, true},
		{"textOutJSON", `{"text":"one", "info":["extra"]}`, true, true},
		{"textOutWrong", `{text:"one", "info":["extra"]}`, false, false},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := printText([]byte(tst.data), tst.jsonFmt)
			if tst.positive {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func Test_printLogin(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		jsonFmt  bool
		positive bool
	}{
		{"loginOut", `{"login":"one","password":"two", "info":["extra"]}`, false, true},
		{"loginOutJSON", `{"login":"one","password":"two", "info":["extra"]}`, true, true},
		{"loginOutWrong", `{login:"one","password":"two", "info":["extra"]}`, false, false},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := printLogin([]byte(tst.data), tst.jsonFmt)
			if tst.positive {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func Test_printBin(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		jsonFmt  bool
		positive bool
	}{
		{"binOut", `{"data":"b25lcGFzc3dvcmR0d28=", "info":["extra"]}`, false, true},
		{"binOutJSON", `{"data":"b25lcGFzc3dvcmR0d28=", "info":["extra"]}`, true, true},
		{"binOutWrong", `{data:"b25lcGFzc3dvcmR0d28=", "info":["extra"]}`, false, false},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := printBin([]byte(tst.data), tst.jsonFmt)
			if tst.positive {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func Test_printCart(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		jsonFmt  bool
		positive bool
	}{
		{"cartOut", `{"number":"one","expired":"two", "holder": "one two", "cvc": "123", "info":["extra"]}`, false, true},
		{"cartOutJSON", `{"number":"one","expired":"two", "holder": "one two", "cvc": "123", "info":["extra"]}`, true, true},
		{"cartOutWrong", `{"number":"one","expired":"two", "holder": "one two", "cvc": 123, "info":["extra"]}`, false, false},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := printCart([]byte(tst.data), tst.jsonFmt)
			if tst.positive {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func Test_printValue(t *testing.T) {
	tests := []struct {
		name     string
		resp     *pb.GetResponse
		jsonFmt  bool
		key      string
		positive bool
	}{
		{"empty", &pb.GetResponse{}, false, "", false},
		{"wrong encrypt", &pb.GetResponse{Key: "one"}, false, "12345", false},
		{"text", &pb.GetResponse{Key: "one", Type: pb.Type_TEXT, Data: []byte("6YLGZ63Cgh52q/deEL/IJQ9rsBUWUPhyBpdSsAN0MMPL3db+/61jvzNDU7G3jkvM9jOVaSDNBTXVIzzV")}, false, "1234567890poiuyt", true},
		{"login", &pb.GetResponse{Key: "one", Type: pb.Type_LOGIN, Data: []byte("KQZI+6vgCmuEGyZl1LIm3NFwNxE42IjYwdjcDXCOgNSep77+Ff62aiJdzx7CwDbMoTX6Adpw76qUld7FMhftjqlvs6P+77G6826BxVZ1")}, false, "1234567890poiuyt", true},
		{"bin", &pb.GetResponse{Key: "one", Type: pb.Type_BINARY, Data: []byte("K6kRNn73QkRK4SFK1miKvJ1vHlVXMgYHu1/Dm0FT3JkKu1yDEBjXtNzlggwKPEBvdkVOjJa3PUEZdqhpbWFFtcXwFAiW06RgvzYKtMg=")}, false, "1234567890poiuyt", true},
		{"cart", &pb.GetResponse{Key: "one", Type: pb.Type_CART, Data: []byte("v8CM6OwP0/MkUhJSdHuN3xf9PaZM0zmy1fXcF/oQ63Vq/WytUwQtARKfgvSYEIOAk1EPp6ZH62lz+jERNciyNbN9FKvIBCqWaE+BEDikmhT1LUuZ1f50Pmgo+wvdZ5hApPmFvkWvn46TCObbKoPogZ8=")}, false, "1234567890poiuyt", true},
	}
	logConfig := zap.NewProductionConfig()
	logger, err := logConfig.Build()
	require.NoError(t, err)
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			if tst.positive {
				require.NoError(t, printValue(tst.resp, tst.jsonFmt, tst.key, logger))
			} else {
				require.Error(t, printValue(tst.resp, tst.jsonFmt, tst.key, logger))
			}
		})
	}
}

// func Test_encrypt(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		data []byte
// 		key  string
// 	}{
// 		{"text", []byte(`{"text":"one", "info":["extra"]}`), "1234567890poiuyt"},
// 		{"bin", []byte(`{"data":"b25lcGFzc3dvcmR0d28=", "info":["extra"]}`), "1234567890poiuyt"},
// 		{"login", []byte(`{"login":"one","password":"two", "info":["extra"]}`), "1234567890poiuyt"},
// 		{"cart", []byte(`{"number":"one","expired":"two", "holder": "one two", "cvc": "123", "info":["extra"]}`), "1234567890poiuyt"},
// 	}
// 	for _, tst := range tests {
// 		t.Run(tst.name, func(t *testing.T) {
// 			out, err := crypto.EncryptKey([]byte(tst.key), tst.data)
// 			require.NoError(t, err)
// 			assert.Equal(t, "", out)
// 		})
// 	}
// }
