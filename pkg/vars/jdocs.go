package vars

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/digital-dream-labs/api/go/jdocspb"
)

var JDOCSDB *sql.DB

// JDOCS need to be accessible by both token and jdocs servers. Token writes a vic.AppTokens jdoc.
func InitJdocsDB(jdocsDB *sql.DB) {
	_, err := jdocsDB.Exec(`
		CREATE TABLE IF NOT EXISTS bot_jdocs (
			thing TEXT NOT NULL,
			name TEXT NOT NULL,
			doc_version INTEGER NOT NULL,
			fmt_version INTEGER NOT NULL,
			client_metadata TEXT NOT NULL,
			json_doc TEXT NOT NULL,
			PRIMARY KEY (thing, name)
		);
	`)
	if err != nil {
		panic("failed to initialize bot_jdocs table: " + err.Error())
	}
	JDOCSDB = jdocsDB
}

func AJdocToJdoc(in AJdoc) jdocspb.Jdoc {
	return jdocspb.Jdoc{
		DocVersion:     in.DocVersion,
		FmtVersion:     in.FmtVersion,
		ClientMetadata: in.ClientMetadata,
		JsonDoc:        in.JsonDoc,
	}
}

type AJdoc struct {
	DocVersion     uint64 `protobuf:"varint,1,opt,name=doc_version,json=docVersion,proto3" json:"doc_version,omitempty"`            // first version = 1; 0 => invalid or doesn't exist
	FmtVersion     uint64 `protobuf:"varint,2,opt,name=fmt_version,json=fmtVersion,proto3" json:"fmt_version,omitempty"`            // first version = 1; 0 => invalid
	ClientMetadata string `protobuf:"bytes,3,opt,name=client_metadata,json=clientMetadata,proto3" json:"client_metadata,omitempty"` // arbitrary client-defined string, eg a data fingerprint (typ "", 32 chars max)
	JsonDoc        string `protobuf:"bytes,4,opt,name=json_doc,json=jsonDoc,proto3" json:"json_doc,omitempty"`
}

type botjdoc struct {
	// vic:00000000
	Thing string `json:"thing"`
	// vic.RobotSettings, etc
	Name string `json:"name"`
	// actual jdoc
	Jdoc AJdoc `json:"jdoc"`
}

func Thingifier(esn string) string {
	esn = strings.ToLower(strings.TrimSpace(esn))
	if strings.HasPrefix(esn, "vic:") {
		return esn
	}
	return "vic:" + esn
}

func WriteJdoc(thing string, name string, jdoc AJdoc) error {
	_, err := JDOCSDB.Exec(
		"INSERT OR REPLACE INTO bot_jdocs (thing, name, doc_version, fmt_version, client_metadata, json_doc) VALUES (?, ?, ?, ?, ?, ?)",
		thing, name, jdoc.DocVersion, jdoc.FmtVersion, jdoc.ClientMetadata, jdoc.JsonDoc,
	)
	if err != nil {
		return errors.New("WriteJdoc: failed to write jdoc: " + err.Error())
	}
	return nil
}

func ReadJdoc(thing string, name string) (AJdoc, error) {
	var jdoc AJdoc
	err := JDOCSDB.QueryRow(
		"SELECT doc_version, fmt_version, client_metadata, json_doc FROM bot_jdocs WHERE thing = ? AND name = ?",
		thing, name,
	).Scan(&jdoc.DocVersion, &jdoc.FmtVersion, &jdoc.ClientMetadata, &jdoc.JsonDoc)
	if err != nil {
		if err == sql.ErrNoRows {
			return AJdoc{}, ErrUserNotFound
		}
		return AJdoc{}, errors.New("ReadJdoc: failed to read jdoc: " + err.Error())
	}
	return jdoc, nil
}
