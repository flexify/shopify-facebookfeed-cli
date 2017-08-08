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
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
	// "github.com/kr/http/limit"
	"github.com/briandowns/spinner"
	"github.com/kr/http/limit"
	"github.com/spf13/cobra"
)

var storeDomain string
var errorMsg []string

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for valid feeds",
	Long:  `Find working limit parameter values and paginated feed urls `,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("store domain required as an argument")
		} else {
			storeDomain = args[0]
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		rss, err := getFeedInfo(getFeedUrl(storeDomain, 0, Limit, NoVariants), context.Background())
		if err != nil {
			log.Fatal(err)
		} else {
			if !rss.Premium {
				fmt.Println("Free User - no pagination available")
			} else {

				if AutoLimit {
					autoRss := recursiveParallelDownload(storeDomain, Limit)
					printFeedUrls(storeDomain, autoRss.PageCount, autoRss.ProductsPerPage)
				} else {
					derr := parallelDownload(storeDomain, rss.PageCount, rss.ProductsPerPage, NoVariants, context.Background())
					if derr == nil {
						printFeedUrls(storeDomain, rss.PageCount, rss.ProductsPerPage)
					} else {
						fmt.Println(RedText("\nSome feed pages did not work."))
						fmt.Println(derr)
						fmt.Println(YellowText("Lower the limit (hint: use the --limit flag) and try again!"))
					}

				}

			}
		}

	},
}

var Limit int
var MaxConcurrentRequests int
var NoVariants bool
var AutoLimit bool

func printFeedUrls(storeDomain string, pageCount int, limit int) {
	fmt.Printf(GreenText("\nA limit of %d works. The Feed urls are:\n"), limit)
	for i := 1; i <= pageCount; i++ {
		pageUrl := getFeedUrl(storeDomain, i, limit, NoVariants).String()
		fmt.Println(pageUrl)
	}
}

func recursiveParallelDownload(storeDomain string, startLimit int) Rss {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // make sure all paths cancel the context to avoid context leak
	fmt.Printf("Testing limit %d \n", startLimit)
	url := getFeedUrl(storeDomain, 0, startLimit, NoVariants)
	rss, err := getFeedInfo(url, ctx)
	if err != nil {
		log.Fatal(err)
	}
	derr := parallelDownload(storeDomain, rss.PageCount, startLimit, NoVariants, ctx)
	cancel()
	if derr == nil {
		return rss
	}

	fmt.Printf(RedText("Limit %d is to high. %s\n"), startLimit, derr)
	// Try new limit parameter
	d := float64(rss.ProductsPerPage) / float64(2)
	newLimit := int(math.Ceil(d))
	return recursiveParallelDownload(storeDomain, newLimit)
}

func init() {

	checkCmd.PersistentFlags().IntVarP(&Limit, "limit", "l", 500, "limit the number of products per feed page")
	checkCmd.PersistentFlags().IntVarP(&MaxConcurrentRequests, "max-concurrent-requests", "c", 10, "limit the number concurrent requests to Shopify's servers")
	checkCmd.PersistentFlags().BoolVarP(&AutoLimit, "auto-limit", "a", false, "find a working limit parameter ")
	checkCmd.PersistentFlags().BoolVarP(&NoVariants, "no-variants", "n", false, "omit all but the first available variant per product")

	RootCmd.AddCommand(checkCmd)
	// Here you will define your flags and configuration settings.

	// Limit to 10 ooncurrent connections per host
	url := func(r *http.Request) interface{} {
		return r.URL.Host
	}
	http.DefaultTransport = &limit.Transport{
		Locker:    limit.By(url, MaxConcurrentRequests),
		Transport: http.DefaultTransport,
	}

}

func getFeedUrl(storeDomain string, page int, limit int, noVariants bool) *url.URL {

	u, err := url.Parse(storeDomain)
	if err != nil {
		log.Fatal(err)
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		u.Host = storeDomain
	}
	u.RawQuery = ""

	q := u.Query()
	if limit != 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if page != 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if noVariants {
		q.Set("no_variants", "")
	}

	u.RawQuery = q.Encode()
	u.Path = "/a/feed/v2/facebook.rss"
	return u
}

func getFeedInfo(feedUrl *url.URL, ctx context.Context) (Rss, error) {
	q := feedUrl.Query()
	q.Set("info", "")
	feedUrl.RawQuery = q.Encode()
	xmlData, err := getContent(feedUrl.String(), ctx)
	var rss Rss
	if err == nil {
		xml.Unmarshal(xmlData, &rss)
	}
	return rss, err
}

func parallelDownload(storeDomain string, pageCount int, limit int, noVariants bool, ctx context.Context) error {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond) // Build our new spinner
	s.Prefix = fmt.Sprintf("Fetching %d products per page (0 of %d pages succeeded) ", limit, pageCount)
	s.Start()
	quit := make(chan bool)
	errc := make(chan error)
	done := make(chan error)
	for i := 1; i <= pageCount; i++ {
		go func(i int) {
			// fetch
			feedUrl := getFeedUrl(storeDomain, i, limit, noVariants)
			//fmt.Println(feedUrl.String())
			_, err := getContent(feedUrl.String(), ctx)
			ch := done // we'll send to done if nil error and to errc otherwise
			if err != nil {
				ch = errc
			}
			select {
			case ch <- err:
				return
			case <-quit:
				return
			}
		}(i)
	}

	count := 0
	for {
		select {
		case err := <-errc:
			close(quit)
			s.Stop()
			return err
		case <-done:
			count++
			s.Prefix = fmt.Sprintf("Fetching %d products per page (%d of %d pages succeeded) ", limit, count, pageCount)
			if count == pageCount {
				s.Stop()
				return nil // got all N signals, so there was no error
			}
		}
	}
}

type Rss struct {
	Premium         bool `xml:"channel>premium"`
	PageCount       int  `xml:"channel>pagecount"`
	ProductsPerPage int  `xml:"channel>products-per-page"`
}
