package common

import (
	"os"
	"strings"
)

type config struct {
	ListenAddress        string
	GinMode              string
	MyDBUrl              string
	MDBUrl               string
	ChroniclesUrl        string
	ChroniclesNamespaces []string
	AccountsUrls         []string
	NewUserKCRole        string
	KCGroupUrl           string
	KmediaKCRole         string
}

func newConfig() *config {
	return &config{
		ListenAddress:        ":8080",
		GinMode:              "debug",
		MyDBUrl:              "postgres://user:password@localhost/mydb?sslmode=disable",
		MDBUrl:               "postgres://user:password@localhost/mdb?sslmode=disable",
		AccountsUrls:         []string{"https://accounts.kab.info/auth/realms/main"},
		ChroniclesUrl:        "https://chronicle-sserver/scan",
		ChroniclesNamespaces: []string{"archive", "kmedia-app-11"},
		NewUserKCRole:        "new_user",
		KmediaKCRole:         "kmedia_user",
	}
}

var Config *config

func Init() {
	Config = newConfig()

	if val := os.Getenv("LISTEN_ADDRESS"); val != "" {
		Config.ListenAddress = val
	}
	if val := os.Getenv("GIN_MODE"); val != "" {
		Config.GinMode = val
	}
	if val := os.Getenv("MYDB_URL"); val != "" {
		Config.MyDBUrl = val
	}
	if val := os.Getenv("MDB_URL"); val != "" {
		Config.MDBUrl = val
	}
	if val := os.Getenv("CHRONICLES_URL"); val != "" {
		Config.ChroniclesUrl = val
	}
	if val := os.Getenv("CHRONICLES_NAMESPACES"); val != "" {
		Config.ChroniclesNamespaces = strings.Split(val, ",")
	}
	if val := os.Getenv("ACCOUNTS_URL"); val != "" {
		Config.AccountsUrls = strings.Split(val, ",")
	}
	if val := os.Getenv("NEW_USER_KC_ROLE"); val != "" {
		Config.NewUserKCRole = val
	}
	if val := os.Getenv("KMEDIA_KC_ROLE"); val != "" {
		Config.KmediaKCRole = val
	}
	if val := os.Getenv("KC_ADD_GROUP_URL"); val != "" {
		Config.KCGroupUrl = val
	}
}
