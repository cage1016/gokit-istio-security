//go:generate $REPO_ROOT/bin/mixer_codegen.sh -a mixer/adapter/authzopa/config/config.proto -x "-s=false -n authzopa-adapter -t authorization"

package authzopa

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"

	stdjwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"istio.io/api/mixer/adapter/model/v1beta1"
	policy "istio.io/api/policy/v1beta1"
	"istio.io/istio/mixer/adapter/authzopa/config"
	"istio.io/istio/mixer/pkg/status"
	"istio.io/istio/mixer/template/authorization"
	"istio.io/pkg/log"

	pb "istio.io/istio/mixer/adapter/authzopa/pb/authz"
)

type (
	// Server is basic server interface
	Server interface {
		Addr() string
		Close() error
		Run(shutdown chan error)
	}

	// AuthzAdapter supports authorization template.
	AuthzAdapter struct {
		listener net.Listener
		server   *grpc.Server
	}
)

var _ authorization.HandleAuthorizationServiceServer = &AuthzAdapter{}

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

func decodeValue(in interface{}) interface{} {
	switch t := in.(type) {
	case *policy.Value_StringValue:
		return t.StringValue
	case *policy.Value_Int64Value:
		return t.Int64Value
	case *policy.Value_DoubleValue:
		return t.DoubleValue
	default:
		return fmt.Sprintf("%v", in)
	}
}

func decodeValueMap(in map[string]*policy.Value) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = decodeValue(v.GetValue())
	}
	return out
}

func (s *AuthzAdapter) HandleAuthorization(ctx context.Context, req *authorization.HandleAuthorizationRequest) (*v1beta1.CheckResult, error) {
	cfg := &config.Params{}

	if req.AdapterConfig != nil {
		if err := cfg.Unmarshal(req.AdapterConfig.Value); err != nil {
			log.Errorf("error unmarshalling adapter config: %v", err)
			return nil, err
		}
	}

	props := decodeValueMap(req.Instance.Subject.Properties)

	// get x jwt payload
	xjp := props["x_jwt_playload"].(string)
	if xjp == "" {
		return &v1beta1.CheckResult{
			Status: status.WithPermissionDenied("Unauthorized: x-jwt-playload Missing"),
		}, nil
	}
	method := props["request_method"].(string)
	path := props["request_path"].(string)

	fmt.Println("xjp", xjp)
	fmt.Println("method", method)

	xp, err := decode([]byte(xjp))
	if err != nil {
		errMsg := fmt.Sprintf("Unauthorized: x-jwt-playload base64 decode fail... %v", err)
		return &v1beta1.CheckResult{
			Status: status.WithPermissionDenied(errMsg),
		}, nil
	}

	var cl stdjwt.MapClaims
	if err = json.NewDecoder(bytes.NewReader(xp)).Decode(&cl); err != nil {
		return &v1beta1.CheckResult{
			Status: status.WithPermissionDenied("Unauthorized: x-jwt-playload claim decode fail"),
		}, nil
	}

	conn, err := grpc.DialContext(ctx, cfg.AuthzServerUrl, grpc.WithInsecure())
	if err != nil {
		return &v1beta1.CheckResult{
			Status: status.WithPermissionDenied("Unauthorized: authz server connect fai"),
		}, nil
	}

	ac := pb.NewAuthzClient(conn)
	res, err := ac.IsAuthorizedReq(context.Background(), &pb.IsAuthorizedReqRequest{
		User:   cl["userId"].(string),
		Path:   path,
		Method: method,
	})
	if err != nil {
		errMsg := fmt.Sprintf("Unauthorized: authz server evaluate fail... %v", err)
		return &v1beta1.CheckResult{
			Status: status.WithPermissionDenied(errMsg),
		}, nil
	}

	fmt.Println("res", res.IsAuthorized)

	if res.IsAuthorized {
		return &v1beta1.CheckResult{
			Status: status.OK,
			// ValidDuration: time.Second * 3,
			// ValidUseCount: 3,
		}, nil
	} else {
		return &v1beta1.CheckResult{
			Status: status.WithPermissionDenied("Unauthorized: permission deny"),
		}, nil
	}
}

// Addr returns the listening address of the server
func (s *AuthzAdapter) Addr() string {
	return s.listener.Addr().String()
}

// Run starts the server run
func (s *AuthzAdapter) Run(shutdown chan error) {
	shutdown <- s.server.Serve(s.listener)
}

// Close gracefully shuts down the server; used for testing
func (s *AuthzAdapter) Close() error {
	if s.server != nil {
		s.server.GracefulStop()
	}

	if s.listener != nil {
		_ = s.listener.Close()
	}

	return nil
}

// NewAuthzAdapter creates a new IBP adapter that listens at provided port.
func NewAuthzAdapter(addr string) (Server, error) {
	if addr == "" {
		addr = "0"
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", addr))
	if err != nil {
		return nil, fmt.Errorf("unable to listen on socket: %v", err)
	}
	s := &AuthzAdapter{
		listener: listener,
	}
	fmt.Printf("listening on \"%v\"\n", s.Addr())
	s.server = grpc.NewServer()
	authorization.RegisterHandleAuthorizationServiceServer(s.server, s)
	return s, nil
}
