package server

import (
	"context"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/hrapovd1/gokeepas/internal/config"
	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/hrapovd1/gokeepas/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestNewKeepPasSrv(t *testing.T) {
	t.Run("right", func(t *testing.T) {
		srv, err := NewKeepPasSrv(&zap.Logger{}, config.Config{DBdsn: "redis://localhost:6379/0"})
		assert.NoError(t, err)
		assert.IsType(t, &KeepPasSrv{}, srv)
		assert.IsType(t, storage.Storage(&storage.RedisStor{}), srv.Stor)
	})

	t.Run("wrong", func(t *testing.T) {
		srv, err := NewKeepPasSrv(&zap.Logger{}, config.Config{DBdsn: ""})
		assert.Error(t, err)
		assert.Nil(t, srv)
	})
}

func TestKeepPasSrv_SignUp(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stor, err := storage.NewRedisStor(config.Config{DBdsn: "redis://localhost/0"})
	require.NoError(t, err)
	storage.SetRedisClient(stor, db)
	srv := KeepPasSrv{Stor: stor, logger: zap.NewNop().Sugar(), conf: config.Config{ServerKey: []byte("wfgxRxAwTILuvwpqD3JSgqnE")}}
	t.Run("right", func(t *testing.T) {
		// expectation
		mock.ExpectHGetAll("/users/test").SetVal(map[string]string{})
		mock.Regexp().ExpectHSet("/users/test", "pass", "ae6d41f07eb6718e95b9cf8a31309e16b0e76c61", "symmkey", `^.*$`, "data", "", "type", "").SetVal(2)
		mock.ExpectHGetAll("/users/test").SetVal(map[string]string{})

		// test
		_, err = srv.SignUp(context.Background(), &pb.AuthRequest{Login: "test", Password: "pass"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("wrong login", func(t *testing.T) {
		_, err = srv.SignUp(context.Background(), &pb.AuthRequest{Login: "server", Password: "pass"})
		assert.Error(t, err)
	})
	t.Run("wrong user name", func(t *testing.T) {
		mock.ExpectHGetAll("/users/test").RedisNil()
		// test
		_, err = srv.SignUp(context.Background(), &pb.AuthRequest{Login: "test", Password: "pass"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("wrong pass", func(t *testing.T) {
		// expectation
		mock.ExpectHGetAll("/users/test").SetVal(map[string]string{"pass": "ae4"})
		mock.ExpectHGetAll("/users/test").SetVal(map[string]string{})

		// test
		_, err = srv.SignUp(context.Background(), &pb.AuthRequest{Login: "test", Password: "pass"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
}

func TestKeepPasSrv_LogIn(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stor, err := storage.NewRedisStor(config.Config{DBdsn: "redis://localhost/0"})
	require.NoError(t, err)
	storage.SetRedisClient(stor, db)
	srv := KeepPasSrv{Stor: stor, logger: zap.NewNop().Sugar(), conf: config.Config{ServerKey: []byte("wfgxRxAwTILuvwpqD3JSgqnE")}}
	t.Run("right", func(t *testing.T) {
		// expectation
		mock.ExpectHGetAll("/users/test").SetVal(map[string]string{})

		_, err = srv.LogIn(context.Background(), &pb.AuthRequest{Login: "test", Password: "pass"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	// "pass", "ae6d41f07eb6718e95b9cf8a31309e16b0e76c61"
	t.Run("get user err", func(t *testing.T) {
		mock.ExpectHGetAll("/users/test").RedisNil()

		_, err = srv.LogIn(context.Background(), &pb.AuthRequest{Login: "test", Password: "pass"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()

	})
	t.Run("wrong pass", func(t *testing.T) {
		mock.ExpectHGetAll("/users/test").SetVal(map[string]string{"pass": "ae6d41f07eb6718e95b9cf8a31309e16b0e76c61"})

		_, err = srv.LogIn(context.Background(), &pb.AuthRequest{Login: "test", Password: "pass"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
}

func TestKeepPasSrv_GetKey(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stor, err := storage.NewRedisStor(config.Config{DBdsn: "redis://localhost/0"})
	require.NoError(t, err)
	storage.SetRedisClient(stor, db)
	srv := KeepPasSrv{Stor: stor, logger: zap.NewNop().Sugar(), conf: config.Config{ServerKey: []byte("wfgxRxAwTILuvwpqD3JSgqnE")}}
	t.Run("right", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"login": "test"}),
		)
		// expectation
		mock.ExpectHGetAll("/users/test").SetVal(map[string]string{"pass": "ae6d41f07eb6718e95b9cf8a31309e16b0e76c61"})

		_, err = srv.GetKey(ctx, &pb.BinRequest{})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("empty metadata", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"logn": "test"}),
		)
		_, err = srv.GetKey(ctx, &pb.BinRequest{})
		assert.Error(t, err)
	})
	t.Run("wrong user", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"login": "test"}),
		)
		// expectation
		mock.ExpectHGetAll("/users/test").SetVal(map[string]string{"pass": ""})

		_, err = srv.GetKey(ctx, &pb.BinRequest{})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestKeepPasSrv_Add(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stor, err := storage.NewRedisStor(config.Config{DBdsn: "redis://localhost/0"})
	require.NoError(t, err)
	storage.SetRedisClient(stor, db)
	srv := KeepPasSrv{Stor: stor, logger: zap.NewNop().Sugar(), conf: config.Config{ServerKey: []byte("wfgxRxAwTILuvwpqD3JSgqnE")}}
	t.Run("right", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"login": "test"}),
		)
		// expectation
		mock.ExpectHSet("test/key", "pass", "", "symmkey", "", "data", "testData", "type", "TEXT").SetVal(2)

		_, err = srv.Add(ctx, &pb.BinRequest{Key: "key", Type: pb.Type_TEXT, Data: "testData"})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("empty login", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"logn": "test"}),
		)

		_, err = srv.Add(ctx, &pb.BinRequest{Key: "key", Type: pb.Type_TEXT, Data: "testData"})
		assert.Error(t, err)
		mock.ClearExpect()
	})
}

func TestKeepPasSrv_Get(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stor, err := storage.NewRedisStor(config.Config{DBdsn: "redis://localhost/0"})
	require.NoError(t, err)
	storage.SetRedisClient(stor, db)
	srv := KeepPasSrv{Stor: stor, logger: zap.NewNop().Sugar(), conf: config.Config{ServerKey: []byte("wfgxRxAwTILuvwpqD3JSgqnE")}}
	t.Run("right", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"login": "test"}),
		)
		// TEXT
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": "TEXT"})
		_, err = srv.Get(ctx, &pb.BinRequest{Key: "key"})
		assert.NoError(t, err)
		// BINARY
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": "BINARY"})
		_, err = srv.Get(ctx, &pb.BinRequest{Key: "key"})
		assert.NoError(t, err)
		// LOGIN
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": "LOGIN"})
		_, err = srv.Get(ctx, &pb.BinRequest{Key: "key"})
		assert.NoError(t, err)
		// CART
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": "CART"})
		_, err = srv.Get(ctx, &pb.BinRequest{Key: "key"})
		assert.NoError(t, err)
		// UNKNOWN
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": ""})
		_, err = srv.Get(ctx, &pb.BinRequest{Key: "key"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("empty login", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"logn": "test"}),
		)

		_, err = srv.Get(ctx, &pb.BinRequest{Key: "key"})
		assert.Error(t, err)
	})
}

func TestKeepPasSrv_Remove(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stor, err := storage.NewRedisStor(config.Config{DBdsn: "redis://localhost/0"})
	require.NoError(t, err)
	storage.SetRedisClient(stor, db)
	srv := KeepPasSrv{Stor: stor, logger: zap.NewNop().Sugar(), conf: config.Config{ServerKey: []byte("wfgxRxAwTILuvwpqD3JSgqnE")}}
	t.Run("right", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"login": "test"}),
		)
		// expectation
		mock.ExpectDel("test/test").SetVal(1)

		_, err = srv.Remove(ctx, &pb.BinRequest{Key: "test"})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("empty login", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"logn": "test"}),
		)
		_, err = srv.Remove(ctx, &pb.BinRequest{Key: "test"})
		assert.Error(t, err)
	})
}

func TestKeepPasSrv_Rename(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stor, err := storage.NewRedisStor(config.Config{DBdsn: "redis://localhost/0"})
	require.NoError(t, err)
	storage.SetRedisClient(stor, db)
	srv := KeepPasSrv{Stor: stor, logger: zap.NewNop().Sugar(), conf: config.Config{ServerKey: []byte("wfgxRxAwTILuvwpqD3JSgqnE")}}
	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.New(map[string]string{"login": "test"}),
	)
	t.Run("right", func(t *testing.T) {
		// expectation
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": "TEXT"})
		mock.ExpectWatch("test/key")
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": "TEXT"})
		mock.ExpectTxPipeline()
		mock.ExpectHSet("test/key1", "pass", "", "symmkey", "", "data", "", "type", "TEXT").SetVal(2)
		mock.ExpectTxPipelineExec()
		mock.ExpectDel("test/key").SetVal(1)

		_, err = srv.Rename(ctx, &pb.BinRequest{Key: "key", NewKey: "key1"})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("get err", func(t *testing.T) {
		mock.ExpectHGetAll("test/key").RedisNil()
		_, err = srv.Rename(ctx, &pb.BinRequest{Key: "key", NewKey: "key1"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("get err2", func(t *testing.T) {
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": ""})
		_, err = srv.Rename(ctx, &pb.BinRequest{Key: "key", NewKey: "key1"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("get err in watch", func(t *testing.T) {
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": "TEXT"})
		mock.ExpectWatch("test/key")
		mock.ExpectHGetAll("test/key").RedisNil()
		_, err = srv.Rename(ctx, &pb.BinRequest{Key: "key", NewKey: "key1"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("remove err", func(t *testing.T) {
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": "TEXT"})
		mock.ExpectWatch("test/key")
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": "TEXT"})
		mock.ExpectTxPipeline()
		mock.ExpectHSet("test/key1", "pass", "", "symmkey", "", "data", "", "type", "TEXT").SetVal(2)
		mock.ExpectTxPipelineExec()
		mock.ExpectDel("test/key").RedisNil()
		_, err = srv.Rename(ctx, &pb.BinRequest{Key: "key", NewKey: "key1"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("empty login", func(t *testing.T) {
		wctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"logn": "test"}),
		)
		_, err = srv.Rename(wctx, &pb.BinRequest{Key: "key", NewKey: "key1"})
		assert.Error(t, err)
	})
}

func TestKeepPasSrv_Update(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stor, err := storage.NewRedisStor(config.Config{DBdsn: "redis://localhost/0"})
	require.NoError(t, err)
	storage.SetRedisClient(stor, db)
	srv := KeepPasSrv{Stor: stor, logger: zap.NewNop().Sugar(), conf: config.Config{ServerKey: []byte("wfgxRxAwTILuvwpqD3JSgqnE")}}
	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.New(map[string]string{"login": "test"}),
	)
	t.Run("right", func(t *testing.T) {
		// expectation
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": "TEXT"})
		mock.ExpectHSet("test/key", "pass", "", "symmkey", "", "data", "", "type", "TEXT").SetVal(2)

		_, err = srv.Update(ctx, &pb.BinRequest{Key: "key"})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("get err", func(t *testing.T) {
		// expectation
		mock.ExpectHGetAll("test/key").RedisNil()

		_, err = srv.Update(ctx, &pb.BinRequest{Key: "key"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("get err", func(t *testing.T) {
		// expectation
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": ""})

		_, err = srv.Update(ctx, &pb.BinRequest{Key: "key"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("update err", func(t *testing.T) {
		// expectation
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": "TEXT"})
		mock.ExpectHSet("test/key", "pass", "", "symmkey", "", "data", "", "type", "TEXT").RedisNil()

		_, err = srv.Update(ctx, &pb.BinRequest{Key: "key"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("empty login", func(t *testing.T) {
		wctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"logn": "test"}),
		)
		_, err = srv.Update(wctx, &pb.BinRequest{Key: "key"})
		assert.Error(t, err)
	})
}

func TestKeepPasSrv_Copy(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stor, err := storage.NewRedisStor(config.Config{DBdsn: "redis://localhost/0"})
	require.NoError(t, err)
	storage.SetRedisClient(stor, db)
	srv := KeepPasSrv{Stor: stor, logger: zap.NewNop().Sugar(), conf: config.Config{ServerKey: []byte("wfgxRxAwTILuvwpqD3JSgqnE")}}
	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.New(map[string]string{"login": "test"}),
	)
	t.Run("right", func(t *testing.T) {
		// expectation
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": "TEXT"})
		mock.ExpectHSet("test/key1", "pass", "", "symmkey", "", "data", "", "type", "TEXT").SetVal(2)

		_, err = srv.Copy(ctx, &pb.BinRequest{Key: "key", NewKey: "key1"})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("get err", func(t *testing.T) {
		// expectation
		mock.ExpectHGetAll("test/key").RedisNil()

		_, err = srv.Copy(ctx, &pb.BinRequest{Key: "key", NewKey: "key1"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("wrong type", func(t *testing.T) {
		// expectation
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": ""})

		_, err = srv.Copy(ctx, &pb.BinRequest{Key: "key", NewKey: "key1"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("copy err", func(t *testing.T) {
		// expectation
		mock.ExpectHGetAll("test/key").SetVal(map[string]string{"data": "", "type": "TEXT"})
		mock.ExpectHSet("test/key1", "pass", "", "symmkey", "", "data", "", "type", "TEXT").RedisNil()

		_, err = srv.Copy(ctx, &pb.BinRequest{Key: "key", NewKey: "key1"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ClearExpect()
	})
	t.Run("empty login", func(t *testing.T) {
		wctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"logn": "test"}),
		)
		_, err = srv.Copy(wctx, &pb.BinRequest{Key: "key", NewKey: "key1"})
		assert.Error(t, err)
	})
}

func TestKeepPasSrv_List(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stor, err := storage.NewRedisStor(config.Config{DBdsn: "redis://localhost/0"})
	require.NoError(t, err)
	storage.SetRedisClient(stor, db)
	srv := KeepPasSrv{Stor: stor, logger: zap.NewNop().Sugar(), conf: config.Config{ServerKey: []byte("wfgxRxAwTILuvwpqD3JSgqnE")}}
	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.New(map[string]string{"login": "test"}),
	)
	// expectation
	mock.ExpectKeys("test/*").SetVal([]string{})

	_, err = srv.List(ctx, &pb.BinRequest{})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestKeepPasSrv_AuthInterceptor(t *testing.T) {
	handler := func(c context.Context, r any) (any, error) {
		return nil, nil
	}
	stor, err := storage.NewRedisStor(config.Config{DBdsn: "redis://localhost/0"})
	require.NoError(t, err)
	srv := KeepPasSrv{Stor: stor, logger: zap.NewNop().Sugar(), conf: config.Config{ServerKey: []byte("wfgxRxAwTILuvwpqD3JSgqnE")}}
	t.Run("auth request", func(t *testing.T) {
		_, err := srv.AuthInterceptor(context.Background(), &pb.AuthRequest{}, &grpc.UnaryServerInfo{}, handler)
		require.NoError(t, err)
	})
	t.Run("empty metadata", func(t *testing.T) {
		_, err := srv.AuthInterceptor(context.Background(), &pb.BinRequest{}, &grpc.UnaryServerInfo{}, handler)
		require.Error(t, err)
	})
	t.Run("wrong token", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"bearer-token": "test"}),
		)
		_, err := srv.AuthInterceptor(ctx, &pb.BinRequest{}, &grpc.UnaryServerInfo{}, handler)
		require.Error(t, err)
	})
}

func TestKeepPasSrv_isValidToken(t *testing.T) {
	srv := KeepPasSrv{conf: config.Config{ServerKey: []byte("wfgxRxAwTILuvwpqD3JSgqnE")}}
	res, err := srv.isValidToken(context.Background(), []string{""})
	assert.Error(t, err)
	assert.Equal(t, "", res)
}
