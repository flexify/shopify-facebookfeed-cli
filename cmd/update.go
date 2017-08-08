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
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/inconshreveable/go-update"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: `Update shopify-facebookfeed to the newest version`,
	Long:  "Update will check for a new release, then if there is an applicable update it will download it and apply it.",
	Run: func(cmd *cobra.Command, args []string) {
		getLatestRelease()
	},
}

var UpdateUrl string

func init() {
	RootCmd.AddCommand(updateCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type Release struct {
	Assets []Asset `json:"assets"`
	Tag    string  `json:"tag_name"`
}

type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func getLatestRelease() {
	resp, err := http.Get("https://api.github.com/repos/flexify/shopify-facebookfeed-cli/releases/latest")
	if err != nil {
		log.Fatal(err)
	}
	jsonData, _ := ioutil.ReadAll(resp.Body)
	var release Release
	json.Unmarshal(jsonData, &release)

	for _, asset := range release.Assets {
		platform := strings.Split(asset.Name, "__")[1]
		if platform == currentPlatform() {
			fmt.Printf("Newest release is %s (%s)\n", release.Tag, currentPlatform())
			err := doUpdate(asset.URL)
			if err != nil {
				fmt.Println(RedText(err))
			}
		}
	}
}

func currentPlatform() string {
	return runtime.GOOS + "-" + runtime.GOARCH
}

func doUpdate(url string) error {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond) // Build our new spinner
	s.Prefix = "Downloading "
	s.Start()
	defer s.Stop()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Status error at %s: %v", url, resp.StatusCode)
	}
	defer resp.Body.Close()
	s.Prefix = "Installing "
	err = update.Apply(resp.Body, update.Options{})
	s.Stop()
	if err != nil {
		if rerr := update.RollbackError(err); rerr != nil {
			return fmt.Errorf("Failed to rollback from bad update: %v", rerr)
		}
		return fmt.Errorf("Could not update and had to roll back. %v", err)
	}
	fmt.Println(GreenText("Newest release installed!"))
	return err
}
