package storage

import (
	"context"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/hrapovd1/gokeepas/internal/config"
	"github.com/hrapovd1/gokeepas/internal/types"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestNewRedisStor(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		positive bool
	}{
		{"right", "redis://localhost:6379/0", true},
		{"wrong", "", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage, err := NewRedisStor(config.Config{
				DBdsn: test.dsn,
			})
			if test.positive {
				require.NoError(t, err)
				assert.IsType(t, &RedisStor{}, storage)
				assert.IsType(t, &redis.Client{}, storage.rdb)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestSetRedisClient(t *testing.T) {
	stor := RedisStor{}
	reds := redis.Client{}
	SetRedisClient(&stor, &reds)
	require.Equal(t, stor.rdb, &reds)
}

func TestRedisStor_Add(t *testing.T) {
	db, mock := redismock.NewClientMock()
	storAdd := RedisStor{rdb: db}
	t.Run("right", func(t *testing.T) {
		mock.ExpectHSet("key", &types.StorageModel{Type: "text", Data: "text"}).SetVal(2)
		err := storAdd.Add(context.Background(), "key", &types.StorageModel{Type: "text", Data: "text"})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("wrong", func(t *testing.T) {
		mock.ExpectHSet("key", &types.StorageModel{Type: "text", Data: "text"}).RedisNil()
		err := storAdd.Add(context.Background(), "key", &types.StorageModel{Type: "text", Data: "text"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
}

func TestRedisStor_Get(t *testing.T) {
	db, mock := redismock.NewClientMock()
	storGet := RedisStor{rdb: db}
	t.Run("right", func(t *testing.T) {
		mock.ExpectHGetAll("key3").SetVal(map[string]string{"type": "text", "data": "text"})
		err := storGet.Get(context.Background(), "key3", &types.StorageModel{Type: "text", Data: "text"})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("wrong", func(t *testing.T) {
		mock.ExpectHGetAll("key3").RedisNil()
		err := storGet.Get(context.Background(), "key3", &types.StorageModel{Type: "text", Data: "text"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()

	})
}

func TestRedisStor_List(t *testing.T) {
	db, mock := redismock.NewClientMock()
	storage := RedisStor{rdb: db}
	mock.ExpectKeys("*").SetVal([]string{"test/1", "test/11"})
	tests := []struct {
		name string
		ctx  context.Context
		in   string
		out  string
	}{
		{
			"two keys",
			metadata.NewIncomingContext(
				context.Background(),
				metadata.New(map[string]string{"login": "test"}),
			),
			"*",
			"'1','11'",
		},
		{
			"without login",
			context.Background(),
			"*",
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := storage.List(test.ctx, test.in)
			assert.Equal(t, test.out, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
	mock.ClearExpect()
	t.Run("one key", func(t *testing.T) {
		mock.ExpectKeys("*").SetVal([]string{"test/1"})
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"login": "test"}))
		result := storage.List(ctx, "*")
		assert.Equal(t, "'1'", result)
	})
	t.Run("special keys", func(t *testing.T) {
		mock.ExpectKeys("*").SetVal([]string{"test/1'1", "test/11"})
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"login": "test"}))
		result := storage.List(ctx, "*")
		assert.Equal(t, `'1\'1','11'`, result)
	})
}

func TestRedisStor_Update(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stor := RedisStor{rdb: db}
	t.Run("right", func(t *testing.T) {
		mock.ExpectHSet("key3", "pass", "", "symmkey", "", "data", "text3", "type", "text").SetVal(2)
		err := stor.Update(context.Background(), "key3", &types.StorageModel{Type: "text", Data: "text3"})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("wrong", func(t *testing.T) {
		mock.ExpectHSet("key3", "pass", "", "symmkey", "", "data", "text3", "type", "text").RedisNil()
		err := stor.Update(context.Background(), "key3", &types.StorageModel{Type: "text", Data: "text3"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRedisStor_Remove(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stor := RedisStor{rdb: db}
	t.Run("right", func(t *testing.T) {
		mock.ExpectDel("key4").SetVal(1)
		err := stor.Remove(context.Background(), "key4")
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("wrong", func(t *testing.T) {
		mock.ExpectDel("key4").RedisNil()
		err := stor.Remove(context.Background(), "key4")
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
}
func TestRedisStor_Copy(t *testing.T) {
	db, mock := redismock.NewClientMock()
	storage := RedisStor{rdb: db}
	mock.ExpectWatch("key")
	mock.ExpectHGetAll("key").SetVal(map[string]string{"type": "text", "data": "text"})
	mock.ExpectTxPipeline()
	mock.ExpectHSet("key1", "pass", "", "symmkey", "", "data", "text", "type", "text").SetVal(2)
	mock.ExpectTxPipelineExec()
	err := storage.Copy(context.Background(), "key", "key1")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ClearExpect()
}

func TestRedisStor_Ping(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stor := RedisStor{rdb: db}
	t.Run("with master key", func(t *testing.T) {
		mock.ExpectPing().SetVal("")
		mock.ExpectHGetAll("server").SetVal(map[string]string{"pass": "d073c221d695acfb17eeb835bf8ec1d4ae8b9655", "symmkey": "", "data": "", "type": ""})
		err := stor.Ping(context.Background(), []byte("test"))
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("wrong ping", func(t *testing.T) {
		mock.ExpectPing().RedisNil()
		err := stor.Ping(context.Background(), []byte("test"))
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("wrong get", func(t *testing.T) {
		mock.ExpectPing().SetVal("")
		mock.ExpectHGetAll("server").RedisNil()
		err := stor.Ping(context.Background(), []byte("test"))
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("empty hash", func(t *testing.T) {
		mock.ExpectPing().SetVal("")
		mock.ExpectHGetAll("server").SetVal(map[string]string{"pass": "", "symmkey": "", "data": "", "type": ""})
		err := stor.Ping(context.Background(), []byte("test"))
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("wrong hash", func(t *testing.T) {
		mock.ExpectPing().SetVal("")
		mock.ExpectHGetAll("server").SetVal(map[string]string{"pass": "c221d695acfb17eeb835bf8ec1d4ae8b9655", "symmkey": "", "data": "", "type": ""})
		err := stor.Ping(context.Background(), []byte("test"))
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
}

func TestRedisStor_Close(t *testing.T) {
	storage, err := NewRedisStor(config.Config{
		DBdsn: "redis://localhost:6379/0",
	})
	require.NoError(t, err)
	err = storage.Close()
	assert.NoError(t, err)
}
