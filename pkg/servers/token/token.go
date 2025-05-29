package token

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cavalier/pkg/sessions"
	"cavalier/pkg/users"
	"cavalier/pkg/vars"

	"github.com/digital-dream-labs/api/go/tokenpb"
	"github.com/google/uuid"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type TokenServer struct {
	tokenpb.UnimplementedTokenServer
}

var (
	TimeFormat     = time.RFC3339Nano
	ExpirationTime = time.Hour * 24
)

func getBotDetailsFromTokReq(ctx context.Context, req *tokenpb.AssociatePrimaryUserRequest) (token string, cert []byte, name string, esn string, err error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", nil, "", "", errors.New("no peer info found in context")
	}
	if p.AuthInfo != nil {
		if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			if len(tlsInfo.State.PeerCertificates) == 0 {
				return "", nil, "", "", errors.New("no peer certificates found")
			}
			clientCert := tlsInfo.State.PeerCertificates[0]
			esn = clientCert.Subject.CommonName
		}
	}
	cert = req.SessionCertificate
	block, _ := pem.Decode(cert)
	certParsed, err := x509.ParseCertificate(block.Bytes)
	if err == nil {
		name = certParsed.Issuer.CommonName
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", nil, "", "", errors.New("no metadata found in context")
	}
	token = md["anki-user-session"][0]
	return token, cert, name, esn, nil
}

func GenJWT(userID, esnThing string) *tokenpb.TokenBundle {
	bundle := &tokenpb.TokenBundle{}
	var tokenJson ClientTokenManager
	guid, tokenHash, _ := CreateTokenAndHashedToken()
	ajdoc, err := vars.ReadJdoc(vars.Thingifier(esnThing), "vic.AppTokens")
	if err != nil {
		ajdoc.DocVersion = 1
		ajdoc.FmtVersion = 1
		ajdoc.ClientMetadata = "wirepod-new-token"
	}
	json.Unmarshal([]byte(ajdoc.JsonDoc), &tokenJson)
	var clientToken ClientToken
	clientToken.IssuedAt = time.Now().Format(TimeFormat)
	clientToken.ClientName = "idontcare"
	clientToken.Hash = tokenHash
	clientToken.AppId = "SDK"
	tokenJson.ClientTokens = append(tokenJson.ClientTokens, clientToken)
	if len(tokenJson.ClientTokens) == 6 {
		finalTokens := tokenJson.ClientTokens[1:]
		tokenJson.ClientTokens = finalTokens
	}
	jdocJsoc, _ := json.Marshal(tokenJson)
	ajdoc.JsonDoc = string(jdocJsoc)
	ajdoc.DocVersion++
	vars.WriteJdoc(vars.Thingifier(esnThing), "vic.AppTokens", ajdoc)
	bundle.ClientToken = guid

	expiresAt := time.Now().AddDate(0, 1, 0).Format(TimeFormat)
	requestUUID := uuid.New().String()
	jwtHeader := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payloadMap := map[string]interface{}{
		"expires":      expiresAt,
		"iat":          time.Now().Format(TimeFormat),
		"permissions":  nil,
		"requestor_id": esnThing,
		"token_id":     requestUUID,
		"token_type":   "user+robot",
		"user_id":      userID,
	}
	payloadBytes, _ := json.Marshal(payloadMap)
	jwtPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	jwtToken := jwtHeader + "." + jwtPayload + "."
	bundle.Token = jwtToken
	return bundle
}

func decodeJWT(tokenString string) (string, string, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) < 2 {
		return "", "", errors.New("invalid token structure")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", fmt.Errorf("payload decode error: %w", err)
	}
	var payload map[string]interface{}
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return "", "", fmt.Errorf("json unmarshal error: %w", err)
	}
	esnThing, ok := payload["requestor_id"].(string)
	if !ok {
		return "", "", errors.New("missing requestor_id")
	}
	userID, ok := payload["user_id"].(string)
	if !ok {
		return "", "", errors.New("missing user_id")
	}
	return esnThing, userID, nil
}

func (s *TokenServer) AssociatePrimaryUser(ctx context.Context, req *tokenpb.AssociatePrimaryUserRequest) (*tokenpb.AssociatePrimaryUserResponse, error) {
	token, cert, name, esn, err := getBotDetailsFromTokReq(ctx, req)
	thing := esn
	esn = strings.TrimPrefix(esn, "vic:")
	if err != nil {
		return nil, err
	}
	if !sessions.IsSessionGood(token) {
		return nil, errors.New("session_expired")
	}
	os.WriteFile(filepath.Join(vars.SessionCertsStorage, name+"_"+esn), cert, 0777)
	bundle := GenJWT(sessions.GetUserIDFromSession(token), thing)
	users.AssociateRobotWithAccount(thing, sessions.GetUserIDFromSession(token))
	return &tokenpb.AssociatePrimaryUserResponse{Data: bundle}, nil
}

func (s *TokenServer) AssociateSecondaryClient(ctx context.Context, req *tokenpb.AssociateSecondaryClientRequest) (*tokenpb.AssociateSecondaryClientResponse, error) {
	token := req.UserSession
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no request metadata")
	}
	jwtToken := md["anki-access-token"]
	thing, userId, err := decodeJWT(jwtToken[0])
	if err != nil {
		return nil, err
	}
	if !users.IsRobotAssociatedWithAccount(thing, userId) {
		return nil, errors.New("bot not associated with account")
	}
	if !sessions.IsSessionGood(token) {
		return nil, errors.New("session_expired")
	}
	bundle := GenJWT(userId, thing)
	return &tokenpb.AssociateSecondaryClientResponse{Data: bundle}, nil
}

func (s *TokenServer) RefreshToken(ctx context.Context, req *tokenpb.RefreshTokenRequest) (*tokenpb.RefreshTokenResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no request metadata")
	}
	jwtToken := md["anki-access-token"]
	thing, userId, err := decodeJWT(jwtToken[0])
	if err != nil {
		return nil, err
	}
	if !users.IsRobotAssociatedWithAccount(thing, userId) {
		return nil, errors.New("bot not associated with account")
	}
	bundle := GenJWT(userId, thing)
	return &tokenpb.RefreshTokenResponse{Data: bundle}, nil
}

func NewTokenServer() *TokenServer {
	return &TokenServer{}
}
