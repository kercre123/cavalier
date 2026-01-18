package cavalier

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

	chipperpb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/api/go/jdocspb"
	"github.com/digital-dream-labs/api/go/tokenpb"
	grpcserver "github.com/digital-dream-labs/hugh/grpc/server"
)

func InitCavalier(InitFunc func() error, SttHandler interface{}, voiceProcessor string) {
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

	certPub, err := os.ReadFile(vars.CertPath)
	if err != nil {
		panic(err)
	}
	certPriv, err := os.ReadFile(vars.KeyPath)
	if err != nil {
		panic(err)
	}
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
	p, err := processreqs.New(InitFunc, SttHandler, voiceProcessor)
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
