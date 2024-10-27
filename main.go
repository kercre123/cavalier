package main

import (
	"cavalier/pkg/servers/accounts"
	"cavalier/pkg/servers/jdocs"
	"cavalier/pkg/servers/token"
	"cavalier/pkg/users"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/digital-dream-labs/api/go/jdocspb"
	"github.com/digital-dream-labs/api/go/tokenpb"
	grpcserver "github.com/digital-dream-labs/hugh/grpc/server"
)

func main() {
	//vars.Init()
	dbConn, err := sql.Open("sqlite3", "./user_database.db")
	if err != nil {
		fmt.Println("Failed to open database connection:", err)
		os.Exit(1)
	}

	defer dbConn.Close()

	users.Init(dbConn)

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
		grpcserver.WithPort(8081),
		grpcserver.WithClientAuth(tls.RequestClientCert),
	//	grpcserver.WithInsecureSkipVerify(),
	)
	if err != nil {
		panic(err)
	}

	// s, _ := chipperserver.New(
	// 	chipperserver.WithIntentProcessor(p),
	// 	chipperserver.WithKnowledgeGraphProcessor(p),
	// 	chipperserver.WithIntentGraphProcessor(p),
	// )

	tokenServer := token.NewTokenServer()
	jdocsServer := jdocs.NewJdocsServer()
	//jdocsserver.IniToJson()

	//chipperpb.RegisterChipperGrpcServer(srv.Transport(), s)
	jdocspb.RegisterJdocsServer(srv.Transport(), jdocsServer)
	tokenpb.RegisterTokenServer(srv.Transport(), tokenServer)

	// listenerOne, err := tls.Listen("tcp", ":8081", &tls.Config{
	// 	Certificates: []tls.Certificate{cert},
	// 	CipherSuites: nil,
	// })
	// if err != nil {
	// 	panic(err)
	// }
	srv.Start()
	http.HandleFunc("/v1", accounts.AccountsAPI)
	http.ListenAndServe(":8080", nil)
}
