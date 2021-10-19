package common

import (
	"fmt"
	"os"
	"strings"
)

type config struct {
	ListenAddress string
	GinMode       string
	MyDBUrl       string
	MDBUrl        string
	ChroniclesUrl string
	AccountsUrls  []string
}

func newConfig() *config {
	return &config{
		ListenAddress: ":8080",
		GinMode:       "debug",
		MyDBUrl:       "postgres://user:password@localhost/mydb?sslmode=disable",
		MDBUrl:        "postgres://user:password@localhost/mdb?sslmode=disable",
		AccountsUrls:  []string{"https://accounts.kab.info/auth/realms/main"},
		ChroniclesUrl: "https://chronicle-sserver/scan",
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
	if val := os.Getenv("ACCOUNTS_URL"); val != "" {
		Config.AccountsUrls = strings.Split(val, ",")
	}

	fmt.Printf("MYDB_URL=%s\n", os.Getenv("MYDB_URL"))
	fmt.Printf("MDB_URL=%s\n", os.Getenv("MDB_URL"))
	fmt.Printf("Config.MyDBUrl=%s\n", Config.MyDBUrl)
	fmt.Printf("Config.MDBUrl=%s\n", Config.MDBUrl)
}
