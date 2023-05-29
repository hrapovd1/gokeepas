// server package contents type and methods of KeepPas server
package server

import (
	"context"

	"github.com/hrapovd1/gokeepas/internal/config"
	"github.com/hrapovd1/gokeepas/internal/crypto"
	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/hrapovd1/gokeepas/internal/storage"
	"github.com/hrapovd1/gokeepas/internal/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// KeepPasSrv type implements grpc server
type KeepPasSrv struct {
	pb.UnimplementedKeepPasServer
	Stor   storage.Storage
	conf   config.Config
	logger *zap.SugaredLogger
}

// NewKeepPasSrv constructs new app grpc server from config
func NewKeepPasSrv(l *zap.Logger, conf config.Config) (*KeepPasSrv, error) {
	storage, err := storage.NewRedisStor(conf)
	if err != nil {
		return nil, err
	}

	server := KeepPasSrv{
		Stor:   storage,
		conf:   conf,
		logger: l.Sugar(),
	}
	return &server, nil
}

// SignUp implements sign up process for new users, it creates new user and makes login for it.
func (kps *KeepPasSrv) SignUp(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	data := types.StorageModel{}
	// check reserved names
	if req.Login == "server" || req.Login == "/users" {
		kps.logger.Debugf("prohibited login: %v", req.Login)
		return nil, status.Error(codes.Unauthenticated, "wrong login or password")
	}
	userKey := "/users/" + req.Login
	if err := kps.Stor.Get(ctx, userKey, &data); err != nil {
		kps.logger.Debug(err)
		return nil, err
	}
	// check if user exists
	if data.PassHash != "" {
		// if exists - log-in
		return kps.LogIn(ctx, req)
	}
	// create new user
	userSymmKey, err := crypto.GenSymmKey(crypto.SymmKeyLength)
	if err != nil {
		kps.logger.Debug(err)
		return nil, err
	}
	data.SymmKey, err = crypto.EncryptKey([]byte(kps.conf.ServerKey), userSymmKey)
	kps.logger.Debugf("srvKey: %v, usrKey: %v", kps.conf.ServerKey, userSymmKey)
	if err != nil {
		kps.logger.Debug(err)
		return nil, err
	}
	data.PassHash, err = crypto.HashPasswd(ctx, []byte(req.Password))
	if err != nil {
		kps.logger.Debug(err)
		return nil, err
	}
	if err := kps.Stor.Add(ctx, userKey, &data); err != nil {
		kps.logger.Debug(err)
		return nil, err
	}
	// log-in after create user
	return kps.LogIn(ctx, req)
}

// LogIn makes login for existed users, it returns bearer token or error
func (kps *KeepPasSrv) LogIn(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	data := types.StorageModel{}
	userKey := "/users/" + req.Login
	if err := kps.Stor.Get(ctx, userKey, &data); err != nil {
		kps.logger.Debug(err)
		return nil, err
	}
	// check if user exists
	if data.PassHash == "" {
		kps.logger.Debugf("got empty pass hash, data: %v", data)
		return nil, status.Error(codes.Unauthenticated, "wrong login or password")
	}
	userToken, err := crypto.GetToken(ctx, req.Login, req.Password, data, kps.conf.ServerKey)
	if err != nil {
		kps.logger.Debug(err)
		return nil, status.Error(codes.Unauthenticated, "wrong login or password")
	}
	symmKey, err := crypto.DecryptKey([]byte(kps.conf.ServerKey), data.SymmKey)
	if err != nil {
		kps.logger.Debug(err)
		return nil, err
	}
	return &pb.AuthResponse{
		SymmKey:   symmKey,
		AuthToken: userToken,
	}, nil
}

// GetKey returns user symmetric key for data encryption
func (kps *KeepPasSrv) GetKey(ctx context.Context, _ *pb.BinRequest) (*pb.AuthResponse, error) {
	data := types.StorageModel{}
	var login string
	md, ok := metadata.FromIncomingContext(ctx)
	kps.logger.Debugf("got metadata: %v", md)
	if ok {
		values := md.Get("login")
		if len(values) > 0 {
			login = values[0]
		} else {
			kps.logger.Debug("Login metadata is empty")
			return nil, status.Error(codes.Unauthenticated, "wrong login or password")
		}
	}
	userKey := "/users/" + login
	if err := kps.Stor.Get(ctx, userKey, &data); err != nil {
		kps.logger.Debug(err)
		return nil, err
	}
	// check if user exists
	if data.PassHash == "" {
		return nil, status.Error(codes.Unauthenticated, "wrong login or password")
	}
	symmKey, err := crypto.DecryptKey([]byte(kps.conf.ServerKey), data.SymmKey)
	if err != nil {
		kps.logger.Debug(err)
		return nil, err
	}
	return &pb.AuthResponse{
		SymmKey: symmKey,
	}, nil
}

// Add implements process of store new secret in storage
func (kps *KeepPasSrv) Add(ctx context.Context, req *pb.BinRequest) (*pb.BinResponse, error) {
	var login string
	md, ok := metadata.FromIncomingContext(ctx)
	kps.logger.Debugf("got metadata: %v", md)
	if ok {
		values := md.Get("login")
		if len(values) > 0 {
			login = values[0]
		} else {
			kps.logger.Debug("Login metadata is empty")
			return nil, status.Error(codes.Unauthenticated, "wrong login or password")
		}
	}
	data := types.StorageModel{Data: string(req.Data), Type: req.Type.String()}
	key := login + "/" + req.Key
	kps.logger.Debugf("name: %v, data: %v", key, req.Data)
	if err := kps.Stor.Add(ctx, key, &data); err != nil {
		kps.logger.Debug(err)
		return nil, err
	}
	return &pb.BinResponse{}, nil
}

// Get implements process of read secret from storage
func (kps *KeepPasSrv) Get(ctx context.Context, req *pb.BinRequest) (*pb.GetResponse, error) {
	var login string
	md, ok := metadata.FromIncomingContext(ctx)
	kps.logger.Debugf("got metadata: %v", md)
	if ok {
		values := md.Get("login")
		if len(values) > 0 {
			login = values[0]
		} else {
			kps.logger.Debug("Login metadata is empty")
			return nil, status.Error(codes.Unauthenticated, "wrong login or password")
		}
	}
	data := types.StorageModel{}
	key := login + "/" + req.Key
	if err := kps.Stor.Get(ctx, key, &data); err != nil {
		kps.logger.Debug(err)
		return nil, err
	}
	resp := pb.GetResponse{Data: []byte(data.Data), Key: req.Key}
	switch data.Type {
	case "TEXT":
		resp.Type = pb.Type_TEXT
	case "BINARY":
		resp.Type = pb.Type_BINARY
	case "LOGIN":
		resp.Type = pb.Type_LOGIN
	case "CART":
		resp.Type = pb.Type_CART
	default:
		return nil, status.Errorf(codes.NotFound, "key doesn't exists")
	}
	return &resp, nil
}

// Remove implements process of delete secret in storage
func (kps *KeepPasSrv) Remove(ctx context.Context, req *pb.BinRequest) (*pb.BinResponse, error) {
	var login string
	md, ok := metadata.FromIncomingContext(ctx)
	kps.logger.Debugf("got metadata: %v", md)
	if ok {
		values := md.Get("login")
		if len(values) > 0 {
			login = values[0]
		} else {
			kps.logger.Debug("Login metadata is empty")
			return nil, status.Error(codes.Unauthenticated, "wrong login or password")
		}
	}
	key := login + "/" + req.Key
	if err := kps.Stor.Remove(ctx, key); err != nil {
		kps.logger.Debug(err)
		return nil, status.Errorf(codes.Internal, "error when remove")
	}
	return &pb.BinResponse{}, nil
}

// Rename implements process of rename existed secret
func (kps *KeepPasSrv) Rename(ctx context.Context, req *pb.BinRequest) (*pb.BinResponse, error) {
	var login string
	md, ok := metadata.FromIncomingContext(ctx)
	kps.logger.Debugf("got metadata: %v", md)
	if ok {
		values := md.Get("login")
		if len(values) > 0 {
			login = values[0]
		} else {
			kps.logger.Debug("Login metadata is empty")
			return nil, status.Error(codes.Unauthenticated, "wrong login or password")
		}
	}
	oldKey := login + "/" + req.Key
	data := types.StorageModel{}
	if err := kps.Stor.Get(ctx, oldKey, &data); err != nil {
		kps.logger.Debug(err)
		return nil, status.Errorf(codes.Internal, "error when rename: %d", err)
	}
	if data.Type == "" {
		return nil, status.Errorf(codes.NotFound, "key doesn't exists")
	}
	newKey := login + "/" + req.NewKey
	if err := kps.Stor.Copy(ctx, oldKey, newKey); err != nil {
		kps.logger.Debug(err)
		return nil, status.Errorf(codes.Internal, "error when rename: %d", err)
	}
	if err := kps.Stor.Remove(ctx, oldKey); err != nil {
		kps.logger.Debug(err)
		return nil, status.Errorf(codes.Internal, "error when rename: %d", err)
	}
	return &pb.BinResponse{}, nil
}

// Update implements update secret in storage without change key
func (kps *KeepPasSrv) Update(ctx context.Context, req *pb.BinRequest) (*pb.BinResponse, error) {
	var login string
	md, ok := metadata.FromIncomingContext(ctx)
	kps.logger.Debugf("got metadata: %v", md)
	if ok {
		values := md.Get("login")
		if len(values) > 0 {
			login = values[0]
		} else {
			kps.logger.Debug("Login metadata is empty")
			return nil, status.Error(codes.Unauthenticated, "wrong login or password")
		}
	}
	key := login + "/" + req.Key
	data := types.StorageModel{}
	if err := kps.Stor.Get(ctx, key, &data); err != nil {
		kps.logger.Debug(err)
		return nil, status.Errorf(codes.Internal, "error when update: %d", err)
	}
	if data.Type == "" {
		return nil, status.Errorf(codes.NotFound, "key doesn't exists")
	}
	if err := kps.Stor.Update(ctx, key, &types.StorageModel{Data: string(req.Data), Type: req.Type.String()}); err != nil {
		kps.logger.Debug(err)
		return nil, status.Errorf(codes.Internal, "error when update: %d", err)
	}
	return &pb.BinResponse{}, nil
}

// Copy implements clone of existed secret
func (kps *KeepPasSrv) Copy(ctx context.Context, req *pb.BinRequest) (*pb.BinResponse, error) {
	var login string
	md, ok := metadata.FromIncomingContext(ctx)
	kps.logger.Debugf("got metadata: %v", md)
	if ok {
		values := md.Get("login")
		if len(values) > 0 {
			login = values[0]
		} else {
			kps.logger.Debug("Login metadata is empty")
			return nil, status.Error(codes.Unauthenticated, "wrong login or password")
		}
	}
	srcKey := login + "/" + req.Key
	data := types.StorageModel{}
	if err := kps.Stor.Get(ctx, srcKey, &data); err != nil {
		kps.logger.Debug(err)
		return nil, status.Errorf(codes.Internal, "error when copy: %d", err)
	}
	if data.Type == "" {
		return nil, status.Errorf(codes.NotFound, "key doesn't exists")
	}
	dstKey := login + "/" + req.NewKey
	if err := kps.Stor.Add(ctx, dstKey, &data); err != nil {
		kps.logger.Debug(err)
		return nil, status.Errorf(codes.Internal, "error when copy: %d", err)
	}
	return &pb.BinResponse{}, nil
}

// List return existed key names in one line: "'key1','key2',..."
func (kps *KeepPasSrv) List(ctx context.Context, _ *pb.BinRequest) (*pb.ListResponse, error) {
	var login string
	md, ok := metadata.FromIncomingContext(ctx)
	kps.logger.Debugf("got metadata: %v", md)
	if ok {
		values := md.Get("login")
		if len(values) > 0 {
			login = values[0]
		} else {
			kps.logger.Debug("Login metadata is empty")
			return nil, status.Error(codes.Unauthenticated, "wrong login or password")
		}
	}
	keyPattern := login + "/*"
	keys := kps.Stor.List(ctx, keyPattern)
	kps.logger.Debugf("List keys: %s", keys)
	return &pb.ListResponse{
		Keys: keys,
	}, nil
}

// AuthInterceptor check bearer token from metadata and allow or reject access
func (kps *KeepPasSrv) AuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if _, ok := req.(*pb.AuthRequest); ok {
		return handler(ctx, req)
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		kps.logger.Debugln("missing metadata")
		return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
	}
	login, err := kps.isValidToken(ctx, md["bearer-token"])
	if err != nil {
		kps.logger.Debugf("check user token error: %v", err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if login == "" {
		kps.logger.Debugf("invalid token, got metadata: %v", md)
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	md.Set("login", login)
	lctx := metadata.NewIncomingContext(ctx, md)

	kps.logger.Debugf("info: %v", info.FullMethod)

	hndlr, err := handler(lctx, req)
	if err != nil {
		kps.logger.Debug(err)
		kps.logger.Errorf("rpc interceptor got error: %v", err)
	}

	return hndlr, err
}

// isValidToken check bearer token
func (kps *KeepPasSrv) isValidToken(_ context.Context, token []string) (string, error) {
	if len(token) == 0 {
		return "", nil
	}
	return crypto.CheckToken(token[0], kps.conf.ServerKey)
}
