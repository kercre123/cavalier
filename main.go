package main

import (
	processreqs "cavalier/pkg/preqs"
	"cavalier/pkg/servers/accounts"
	chipperserver "cavalier/pkg/servers/chipper"
	"cavalier/pkg/servers/jdocs"
	"cavalier/pkg/servers/token"
	"cavalier/pkg/sessions"
	"cavalier/pkg/users"
	"cavalier/pkg/vars"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"

	stt "cavalier/pkg/vosk"

	chipperpb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/api/go/jdocspb"
	"github.com/digital-dream-labs/api/go/tokenpb"
	grpcserver "github.com/digital-dream-labs/hugh/grpc/server"
)

func main() {
	vars.Init()
	dbConn, err := sql.Open("sqlite3", "./user_database.db")
	if err != nil {
		fmt.Println("Failed to open database connection:", err)
		os.Exit(1)
	}

	defer dbConn.Close()

	dbConnJdocs, err := sql.Open("sqlite3", "./bot_database.db")
	if err != nil {
		fmt.Println("Failed to open jdocs database connection:", err)
		os.Exit(1)
	}

	defer dbConn.Close()
	defer dbConnJdocs.Close()

	users.Init(dbConn)
	vars.InitJdocsDB(dbConnJdocs)
	sessions.Init()

	certPub, _ := os.ReadFile("./cert.crt")
	certPriv, _ := os.ReadFile("./cert.key")
	cert, err := tls.X509KeyPair(certPub, certPriv)
	if err != nil {
		panic(err)
	}

	srv, err := grpcserver.New(
		grpcserver.WithViper(),
		grpcserver.WithReflectionService(),
		grpcserver.WithCertificate(cert),
		grpcserver.WithClientAuth(tls.RequestClientCert),
	//	grpcserver.WithInsecureSkipVerify(),
	)
	if err != nil {
		panic(err)
	}
	p, err := processreqs.New(stt.Init, stt.STT, stt.Name)
	if err != nil {
		panic(err)
	}
	s, _ := chipperserver.New(
		chipperserver.WithIntentProcessor(p),
		chipperserver.WithKnowledgeGraphProcessor(p),
		chipperserver.WithIntentGraphProcessor(p),
	)

	tokenServer := token.NewTokenServer()
	jdocsServer := jdocs.NewJdocsServer()
	//jdocsserver.IniToJson()

	chipperpb.RegisterChipperGrpcServer(srv.Transport(), s)
	jdocspb.RegisterJdocsServer(srv.Transport(), jdocsServer)
	tokenpb.RegisterTokenServer(srv.Transport(), tokenServer)

	listenerOne, err := net.Listen("tcp", ":8081")
	if err != nil {
		panic(err)
	}
	go srv.Transport().Serve(listenerOne)
	http.HandleFunc("/v1/", accounts.AccountsAPI)
	http.ListenAndServe(":8080", nil)
}
