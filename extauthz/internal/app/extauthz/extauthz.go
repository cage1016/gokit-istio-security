package extauthz

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	stdjwt "github.com/dgrijalva/jwt-go"
	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gogo/googleapis/google/rpc"
	pb "github.com/qeek-dev/ext-authz/pb/authz"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
)

func decode(enc []byte) ([]byte, error) {
	// create new buffer from enc
	// you can also use bytes.NewBuffer(enc)
	r := bytes.NewReader(enc)
	// pass it to NewDecoder so that it can read data
	dec := base64.NewDecoder(base64.RawURLEncoding, r)
	// read decoded data from dec to res
	res, err := ioutil.ReadAll(dec)
	return res, err
}

type AuthorizationServer struct {
	logger log.Logger
	ac     pb.AuthzClient
	byPass bool
}

func NewAuthorizationServer(conn *grpc.ClientConn, byPass bool, logger log.Logger) AuthorizationServer {
	return AuthorizationServer{
		ac:     pb.NewAuthzClient(conn),
		byPass: byPass,
		logger: logger,
	}
}

func (as *AuthorizationServer) Check(ctx context.Context, req *auth.CheckRequest) (*auth.CheckResponse, error) {
	// level.Info(as.logger).Log("msg", fmt.Sprintf("gRPC check ok attributes: %s\n", req.Attributes))
	h := req.GetAttributes().GetRequest().GetHttp()

	level.Info(as.logger).Log("x-envoy-original-path", h.GetHeaders()["x-envoy-original-path"])
	level.Info(as.logger).Log("path", h.Path)

	if as.byPass {
		level.Info(as.logger).Log("byPass", as.byPass)
		return &auth.CheckResponse{
			Status: &rpcstatus.Status{
				Code: int32(rpc.OK),
			},
		}, nil
	}

	// TODO: workaround bypass for API: login & any GRPC request, should exclude those request by envoy filter
	if strings.HasSuffix(h.GetHeaders()["x-envoy-original-path"], "login") || strings.HasPrefix(h.Path, "/pb.") {
		return &auth.CheckResponse{
			Status: &rpcstatus.Status{
				Code: int32(rpc.OK),
			},
		}, nil
	}

	s := as.Verify(ctx, h.GetHeaders()["x-envoy-original-path"], h.Method, h.GetHeaders()["x-jwt-playload"])
	return &auth.CheckResponse{
		Status: s,
	}, nil
}

func (as *AuthorizationServer) Verify(ctx context.Context, path, method, xjwt string) *rpcstatus.Status {
	xp, _ := decode([]byte(xjwt))

	var cl stdjwt.MapClaims
	if err := json.NewDecoder(bytes.NewReader(xp)).Decode(&cl); err != nil {
		level.Error(as.logger).Log("msg", "stdjwt decode err", "err", err)
		return &rpcstatus.Status{
			Message: "Unauthorized: authz server connect fai",
			Code:    int32(rpc.PERMISSION_DENIED),
		}
	}

	res, err := as.ac.IsAuthorizedReq(ctx, &pb.IsAuthorizedReqRequest{
		User:   cl["userId"].(string),
		Path:   path,
		Method: method,
	})

	if err != nil {
		level.Error(as.logger).Log("err", err)
		return &rpcstatus.Status{
			Message: fmt.Sprintf("Unauthorized: authz server evaluate fail... %v", err),
			Code:    int32(rpc.PERMISSION_DENIED),
		}
	}

	as.logger.Log("user", cl["userId"].(string), "path", path, "method", method, "res", res.IsAuthorized)
	if res.IsAuthorized == false {
		return &rpcstatus.Status{
			Message: "Unauthorized: permission denied",
			Code:    int32(rpc.PERMISSION_DENIED),
		}
	}

	return &rpcstatus.Status{
		Code: int32(rpc.OK),
	}
}
