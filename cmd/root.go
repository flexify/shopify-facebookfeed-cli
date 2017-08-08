// Copyright © 2017 flexify.net
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	. "github.com/tj/go-debug"
)

const rowSeparator string = "<!--|row|-->"
const colSeparator string = "<!--|col|-->"

// init global flags
var cfgFile, outputFile string
var debug = Debug("cli")

var RootCmd = &cobra.Command{
	Use:   "shopify-facebookfeed",
	Short: "shopify-facebookfeed is a tool to check and find a valid facebook feed configuration.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

/* colors */

// RedText is a func that wraps a string in red color tags and it will be
// red when printed out.
var RedText = color.New(color.FgRed).SprintFunc()

// YellowText is a func that wraps a string in yellow color tags and it will be
// yellow when printed out.
var YellowText = color.New(color.FgYellow).SprintFunc()

// BlueText is a func that wraps a string in blue color tags and it will be
// blue when printed out.
var BlueText = color.New(color.FgBlue).SprintFunc()

// GreenText is a func that wraps a string in green color tags and it will be
// green when printed out.
var GreenText = color.New(color.FgGreen).SprintFunc()

// CyanText is a func that wraps a string in cyan color tags and it will be
// cyan when printed out.
var CyanText = color.New(color.FgCyan).SprintFunc()

func getContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Read body: %v", err)
	}
	responseString := string(data)
	if responseString == "Liquid error: Memory limits exceeded" {
		return nil, fmt.Errorf("Memory error at %s", url)
	}

	return data, nil
}

func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}
