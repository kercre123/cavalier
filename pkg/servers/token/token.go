package token

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
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

	ijwt "github.com/dgrijalva/jwt-go"
	"github.com/digital-dream-labs/api/go/tokenpb"
	"github.com/golang-jwt/jwt"
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

// returns session token, session cert, robot name ("Vector-####"), then thing ("vic:esn")
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

	// get metadata
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
	var finalTokens []ClientToken
	// limit tokens to 6, don't fill the db
	if len(tokenJson.ClientTokens) == 6 {
		for i, tok := range tokenJson.ClientTokens {
			if i != 0 {
				finalTokens = append(finalTokens, tok)
			}
		}
		tokenJson.ClientTokens = finalTokens
	}
	jdocJsoc, err := json.Marshal(tokenJson)
	ajdoc.JsonDoc = string(jdocJsoc)
	ajdoc.DocVersion++
	vars.WriteJdoc(vars.Thingifier(esnThing), "vic.AppTokens", ajdoc)

	bundle.ClientToken = guid

	currentTime := time.Now().Format(TimeFormat)
	expiresAt := time.Now().AddDate(0, 1, 0).Format(TimeFormat)
	fmt.Println("Current time: " + currentTime)
	fmt.Println("Token expires: " + expiresAt)
	requestUUID := uuid.New().String()
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"expires":      expiresAt,
		"iat":          currentTime,
		"permissions":  nil,
		"requestor_id": esnThing,
		"token_id":     requestUUID,
		"token_type":   "user+robot",
		"user_id":      userID,
	})
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	tokenString, _ := token.SignedString(rsaKey)
	bundle.Token = tokenString
	return bundle
}

func decodeJWT(tokenString string) (string, string, error) {
	token, _, err := new(ijwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", "", fmt.Errorf("error parsing token: %w", err)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		esnThing := claims["requestor_id"]
		userID := claims["user_id"]
		esnThingStr, ok := esnThing.(string)
		if !ok {
			return "", "", errors.New("token does not have an esn")
		}
		userIDStr, ok := userID.(string)
		if !ok {
			return "", "", errors.New("token does not have a user id")
		}
		return esnThingStr, userIDStr, nil
	}

	return "", "", errors.New("invalid token or claims")
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
	return &tokenpb.AssociatePrimaryUserResponse{
		Data: bundle,
	}, nil
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
	return &tokenpb.AssociateSecondaryClientResponse{
		Data: bundle,
	}, nil
}

// INSECURE!
// i don't have a way to verify the incoming JWT, unless i save the generated key from the primary request.. that's an idea
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
	return &tokenpb.RefreshTokenResponse{
		Data: bundle,
	}, nil
}

func NewTokenServer() *TokenServer {
	return &TokenServer{}
}
