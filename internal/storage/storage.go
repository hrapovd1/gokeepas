/*
Package storage contents types and methods for KeepPas server storage
*/
package storage

import (
	"context"
	"errors"
	"strings"

	"github.com/hrapovd1/gokeepas/internal/config"
	"github.com/hrapovd1/gokeepas/internal/crypto"
	"github.com/hrapovd1/gokeepas/internal/types"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/metadata"
)

const transactWatchRetries = 100 // count of retries of transaction in Copy method

type Storage interface {
	Add(context.Context, string, *types.StorageModel) error
	Get(context.Context, string, *types.StorageModel) error
	Remove(context.Context, string) error
	Update(context.Context, string, *types.StorageModel) error
	Copy(context.Context, string, string) error
	Ping(context.Context, []byte) error
	List(context.Context, string) string
	Close() error
}

// RedisStor type implements os Storage interface. It uses redis db as back storage.
type RedisStor struct {
	rdb *redis.Client
}

// NewRedisStor creates new RedisStor according server configuration
func NewRedisStor(conf config.Config) (*RedisStor, error) {
	redisOpt, err := redis.ParseURL(conf.DBdsn)
	if err != nil {
		return nil, err
	}
	rs := RedisStor{
		rdb: redis.NewClient(redisOpt),
	}
	return &rs, nil
}

// SetRedisClient set redis client in storage, use in tests.
func SetRedisClient(rs *RedisStor, rc *redis.Client) {
	rs.rdb = rc
}

// Add implements add process key/value in storage.
func (rs RedisStor) Add(ctx context.Context, key string, val *types.StorageModel) error {
	return rs.rdb.HSet(ctx, key, val).Err()
}

// Get returns key/value from storage.
func (rs RedisStor) Get(ctx context.Context, key string, val *types.StorageModel) error {
	return rs.rdb.HGetAll(ctx, key).Scan(val)
}

// List reads all existed user's keys
func (rs RedisStor) List(ctx context.Context, pattern string) string {
	res := rs.rdb.Keys(ctx, pattern)
	out := ""
	keys := res.Val()
	commaCount := len(keys) - 1
	if commaCount+1 == 0 {
		return out
	}
	var login string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get("login")
		if len(values) > 0 {
			login = values[0]
		} else {
			return out
		}
	}
	prefixLen := len(login + "/")
	// join keys in one line: "'key1','key2'..."
	for _, val := range keys {
		select {
		case <-ctx.Done():
			return out
		default:
			// process each key from query
			tmpVal := `'`
			val = val[prefixLen:]
			// check if key contents ' - escape it
			tmpVals := strings.Split(val, "'")
			if len(tmpVals) > 1 {
				escapeCount := len(tmpVals) - 1
				for _, v := range tmpVals {
					tmpVal += v
					if escapeCount > 0 {
						tmpVal += `\'`
						escapeCount--
					}
				}
				tmpVal += `'`
			} else {
				tmpVal += val + `'`
			}
			out += tmpVal
			if commaCount > 0 {
				out += `,`
				commaCount--
			}
		}
	}
	return out
}

// Update change existed key/value in storage.
func (rs RedisStor) Update(ctx context.Context, key string, val *types.StorageModel) error {
	return rs.rdb.HSet(ctx, key, val).Err()
}

// Remove remove existed key in storage
func (rs RedisStor) Remove(ctx context.Context, key string) error {
	return rs.rdb.Del(ctx, key).Err()
}

// Copy clones existed key/value in new key/value.
func (rs RedisStor) Copy(ctx context.Context, srcKey string, dstKey string) error {
	values := types.StorageModel{}
	var err error
	txf := func(tx *redis.Tx) error {
		// Get the current value or zero.
		if err = tx.HGetAll(ctx, srcKey).Scan(&values); err != nil {
			return err
		}

		// Operation is commited only if the watched keys remain unchanged.
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			return pipe.HSet(ctx, dstKey, &values).Err()
		})
		return err
	}
	// Retry if the key has been changed.
	for i := 0; i < transactWatchRetries; i++ {
		err := rs.rdb.Watch(ctx, txf, srcKey)
		if err == nil {
			// Success.
			return nil
		}
		if err == redis.TxFailedErr {
			// Optimistic lock lost. Retry.
			continue
		}
		// Return any other error.
		return err
	}

	return errors.New("increment reached maximum number of retries")
}

// Ping check connection to storage and check server master key hash in storage.
// If hash exists, it will be compared with server master key from server configuration.
// Else new hash will be created and saved in storage.
func (rs RedisStor) Ping(ctx context.Context, srvKey []byte) error {
	if err := rs.rdb.Ping(ctx).Err(); err != nil {
		return err
	}
	data := types.StorageModel{}
	if err := rs.Get(ctx, "server", &data); err != nil {
		return err
	}
	if data.PassHash == "" {
		srvHash, err := crypto.HashPasswd(ctx, srvKey)
		if err != nil {
			return err
		}
		data.PassHash = srvHash
		if err := rs.Add(ctx, "server", &data); err != nil {
			return err
		}
	} else {
		srvHash, err := crypto.HashPasswd(ctx, srvKey)
		if err != nil {
			return err
		}
		if data.PassHash != srvHash {
			return errors.New("server hash in db doesn't match! Use different db")
		}
	}
	return nil
}

// Close closes connection to storage db.
func (rs RedisStor) Close() error {
	return rs.rdb.Close()
}
