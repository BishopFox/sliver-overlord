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
	"os"

	"github.com/spf13/cobra"
)

const (
	// ANSI Colors
	normal    = "\033[0m"
	black     = "\033[30m"
	red       = "\033[31m"
	green     = "\033[32m"
	orange    = "\033[33m"
	blue      = "\033[34m"
	purple    = "\033[35m"
	cyan      = "\033[36m"
	gray      = "\033[37m"
	bold      = "\033[1m"
	clearln   = "\r\x1b[2K"
	upN       = "\033[%dA"
	downN     = "\033[%dB"
	underline = "\033[4m"

	// Info - Display colorful information
	Info = bold + cyan + "[*] " + normal
	// Warn - Warn a user
	Warn = bold + red + "[!] " + normal
	// Debug - Display debug information
	Debug = bold + purple + "[-] " + normal
	// Woot - Display success
	Woot = bold + green + "[$] " + normal

	// *** CLI Flags ***

	// Enum flags
	remoteDebuggingPortFlagStr = "remote-debugging-port"

	// Inject flags
	extensionIDStrFlag = "extension-id"
	wsURLFlagStr       = "ws-url"
	jsCodeURLFlagStr   = "js-url"
	jsCodeFlagStr      = "js-code"

	// Screenshot flags
	targetIDStrFlag = "target-id"

	// Peristent flags
	outputFormatFlag = "format"
	verboseFlagStr   = "verbose"

	// Output formats
	consoleOutput = "console"
	jsonOutput    = "json"

	// *** Process Exit Codes ***

	// ExitSuccess (0) - Success
	ExitSuccess = iota
	// ExitRootCmdError (1) - Root command returned error
	ExitRootCmdError
	// ExitBadFlag (2) - Bad CLI flag
	ExitBadFlag
	// ExitNoJSPayload (3) - Failed to find JS payload
	ExitNoJSPayload
	// ExitDebugQueryError (4) - Failed to query remote debugging API
	ExitDebugQueryError
	// ExitExecuteJSError (5) - Evaluation of the JS payload returned an error
	ExitExecuteJSError
	// ExitTargetNotFound (6) - No valid injection target found
	ExitTargetNotFound
	// ExitMarshalingErr (7) - A marshalling error occurred
	ExitMarshalingErr
	// ExitTaskFailure (8) - A chromedp task failed
	ExitTaskFailure
)

var (
	outputFormats = []string{consoleOutput, jsonOutput}
)

func init() {

	// Cursed
	curseCmd.Flags().IntP(remoteDebuggingPortFlagStr, "r", 1099, "remote debugging port")
	curseCmd.Flags().StringP(jsCodeURLFlagStr, "j", "", "js code url")
	curseCmd.Flags().StringP(jsCodeFlagStr, "J", "", "js code")
	rootCmd.AddCommand(curseCmd)

	// Enum
	enumCmd.Flags().IntP(remoteDebuggingPortFlagStr, "r", 1099, "remote debugging port")
	rootCmd.AddCommand(enumCmd)

	// Manifest
	manifestCmd.Flags().IntP(remoteDebuggingPortFlagStr, "r", 1099, "remote debugging port")
	manifestCmd.Flags().StringP(extensionIDStrFlag, "e", "", "extension id")
	rootCmd.AddCommand(manifestCmd)

	// Inject
	injectCmd.Flags().IntP(remoteDebuggingPortFlagStr, "r", 1099, "remote debugging port")
	injectCmd.Flags().StringP(extensionIDStrFlag, "e", "", "extension id")
	injectCmd.Flags().StringP(jsCodeURLFlagStr, "j", "", "js code url")
	injectCmd.Flags().StringP(jsCodeFlagStr, "J", "", "js code")
	rootCmd.AddCommand(injectCmd)

	// Screenshot
	screenshotCmd.Flags().IntP(remoteDebuggingPortFlagStr, "r", 1099, "remote debugging port")
	screenshotCmd.Flags().StringP(targetIDStrFlag, "t", "", "target id")
	rootCmd.AddCommand(screenshotCmd)

	// Root
	rootCmd.PersistentFlags().BoolP(verboseFlagStr, "V", false, "verbose output")
	rootCmd.PersistentFlags().StringP(outputFormatFlag, "f", consoleOutput, fmt.Sprintf("output format: %v", outputFormats))
}

var rootCmd = &cobra.Command{
	Use:   "overlord",
	Short: "Chrome Extension and Electron code injection tool",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(ExitSuccess)
	},
}

// Execute - Execute root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(ExitRootCmdError)
	}
}

func getOutputFlags(cmd *cobra.Command) (string, bool) {
	verbose, err := cmd.Flags().GetBool(verboseFlagStr)
	if err != nil {
		fmt.Printf(Warn+"Failed to parse --%s flag: %s\n", verboseFlagStr, err)
		os.Exit(ExitBadFlag)
	}
	format, err := cmd.Flags().GetString(outputFormatFlag)
	if err != nil {
		fmt.Printf(Warn+"Failed to parse --%s flag: %s\n", outputFormatFlag, err)
		os.Exit(ExitBadFlag)
	}
	if !contains(outputFormats, format) {
		fmt.Printf(Warn+"Invalid output format %s\n", format)
		os.Exit(ExitBadFlag)
	}
	return format, verbose
}

func contains(haystack []string, needle string) bool {
	set := make(map[string]struct{}, len(haystack))
	for _, entry := range haystack {
		set[entry] = struct{}{}
	}
	_, ok := set[needle]
	return ok
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

func getJSCode(cmd *cobra.Command) string {
	jsCode, err := cmd.Flags().GetString(jsCodeFlagStr)
	if err != nil {
		fmt.Printf(Warn+"Failed to parse --%s flag: %s\n", jsCodeFlagStr, err)
		os.Exit(ExitBadFlag)
	}
	if jsCode != "" {
		return jsCode
	}

	jsURL, err := cmd.Flags().GetString(jsCodeURLFlagStr)
	if err != nil {
		fmt.Printf(Warn+"Failed to parse --%s flag: %s\n", jsCodeURLFlagStr, err)
		os.Exit(ExitBadFlag)
	}
	if jsURL == "" && jsCode == "" {
		fmt.Printf(Warn+"You must specify --%s or --%s\n", jsCodeFlagStr, jsCodeURLFlagStr)
		os.Exit(ExitNoJSPayload)
	}

	jsCode, err = fetchJSCode(jsURL)
	if err != nil {
		fmt.Printf(Warn+"Failed to fetch JS code from url: %s\n", err)
		os.Exit(ExitNoJSPayload)
	}
	if jsCode == "" {
		fmt.Println(Warn + "No JS payload, see --help")
		os.Exit(ExitNoJSPayload)
	}
	return jsCode
}
