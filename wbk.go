/*
** name: wbk (waybackurls)
** cred: https://github.com/tomnomnom/hacks/tree/master/waybackurls
** docs: https://github.com/internetarchive/wayback/tree/master/wayback-cdx-server#url-match-scope
** m4ll0k (github.com/m4ll0k)

DOC LINK: https://bit.ly/2UCEIyH

main features compared to gau / waybackurls:

filtring (https://github.com/internetarchive/wayback/tree/master/wayback-cdx-server#filtering):
  Date Range:
  E.g:
	  $ wbk -fromto "2010-2015" paypal.com

  Regex filtering
  E.g:
	  $ wbk -filter "statuscode:200,mimetype:application/json" paypal.com     # show only urls with statuscode 200 and mimetype application/json
	  $ wbk -filter "\!statuscode:200,\!mimetype:application/json" paypal.com # not show urls with statuscode 200 and mimetype application/json

url match scope (https://github.com/internetarchive/wayback/tree/master/wayback-cdx-server#url-match-scope)
	E.g:
	  $ wbk -match "exact" paypal.com/signin
	  $ wbk -match "host"  paypal.com
	  $ ...

inputs
	E.g:
		$ cat mydomains.txt | wbk ... # stdin
		$ wbk ... mydomains.txt       # file
		$ wbk ... mydomain.com        # string


*/


package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "strings"
)


const (
	Version      = "v0.1 (beta)"
	FetchCDXBase = "http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=json&fl=original&collapse=urlkey"
)

var (
	fromToFlag  = flag.String("fromto", "", "results filtered by timestamp using from= and to= params\nE.g: -fromto \"2012-2015\"")
	filterFlag  = flag.String("filter", "", "filter:\n\tmimetype:\tE.g: mimetype:application/json\n\tstatuscode:\tE.g: statuscode:200\nE.g: -filter \"mimetype:application/json\"")
	matchFlag   = flag.String("match", "", "exact\n\tprefix\n\thost\n\tdomain")
	versionFlag = flag.Bool("version", false, "show version and exit\n\nDoc link: https://bit.ly/2UCEIyH")
)

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Println("WBK version:", Version)
		return
	}

	if flag.NArg() < 1 {
		fmt.Println("Usage: wbk [options] <domain>")
		flag.PrintDefaults()
		return
	}

	domain := flag.Arg(0)

	fetchCDX := buildFetchCDXURL(domain)

	urls, err := fetchURLs(fetchCDX)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch URLs for domain %s: %v\n", domain, err)
		return
	}

	for _, url := range urls {
		fmt.Println(url)
	}
}

func buildFetchCDXURL(domain string) string {
	fetchCDX := FetchCDXBase

	if strings.Contains(*fromToFlag, "-") {
		fetchCDX += "&from=" + strings.Split(*fromToFlag, "-")[0] + "&to=" + strings.Split(*fromToFlag, "-")[1]
	}

	if *filterFlag != "" {
		filters := strings.Split(*filterFlag, ",")
		for _, v := range filters {
			filter := strings.Split(v, ":")[0]
			if filter == "statuscode" || filter == "mimetype" || filter == "!statuscode" || filter == "!mimetype" {
				fetchCDX += "&filter=" + v
			}
		}
	}

	if *matchFlag != "" {
		fetchCDX += "&matchType=" + *matchFlag
	}

	return fmt.Sprintf(fetchCDX, domain)
}

func fetchURLs(fetchCDX string) ([]string, error) {
	response, err := http.Get(fetchCDX)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var urls [][]string
	if err := json.Unmarshal(body, &urls); err != nil {
		return nil, err
	}

	var result []string
	for _, url := range urls {
		if len(url) > 0 {
			result = append(result, url[0])
		}
	}

	return result, nil
}
