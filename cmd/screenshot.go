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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/bishopfox/sliver-overlord/pkg/overlord"
	"github.com/spf13/cobra"
)

// Screenshot - Marshal screenshot
type Screenshot struct {
	PNG string `json:"png"`
}

var screenshotCmd = &cobra.Command{
	Use:   "screenshot",
	Short: "Screenshot a Chrome context",
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
		targetID, err := cmd.Flags().GetString(targetIDStrFlag)
		if err != nil {
			fmt.Printf(Warn+"Failed to parse --%s flag: %s\n", targetIDStrFlag, err)
			os.Exit(ExitBadFlag)
		}
		format, _ := getOutputFlags(cmd)

		debugURL := url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("localhost:%d", debuggingPort),
			Path:   "/json",
		}

		targets, err := overlord.QueryDebugTargets(debugURL.String())
		if err != nil {
			fmt.Printf(Warn+"%s\n", err)
			os.Exit(ExitDebugQueryError)
		}

		found := false
		result := []byte{}
		for _, target := range targets {
			if target.ID == targetID {
				fmt.Printf(Info + "Waiting for repaint ...\n")
				result, err = overlord.Screenshot(target.WebSocketDebuggerURL, targetID, 100)
				if err != nil {
					fmt.Printf(Warn+"Failed to take screenshot %s\n", err)
					os.Exit(ExitTaskFailure)
				}
				found = true
				break
			}
		}
		if !found {
			fmt.Printf(Warn+"Target '%s' not found\n", targetID)
			os.Exit(ExitTargetNotFound)
		}

		if format == consoleOutput {
			save, err := cmd.Flags().GetString(saveFlagStr)
			if err != nil {
				fmt.Printf(Warn+"Failed to parse --%s flag: %s\n", saveFlagStr, err)
				os.Exit(ExitBadFlag)
			}
			if err := ioutil.WriteFile(save, result, 0o644); err != nil {
				fmt.Printf(Warn+"File write failed %s\n", err)
				os.Exit(ExitFileIOError)
			}
		}

		if format == jsonOutput {
			data, err := json.Marshal(&Screenshot{PNG: base64.StdEncoding.EncodeToString(result)})
			if err != nil {
				fmt.Printf(Warn+"Failed to marshal json %s\n", err)
				os.Exit(ExitMarshalingErr)
			}
			fmt.Printf(string(data))
		}
	},
}
