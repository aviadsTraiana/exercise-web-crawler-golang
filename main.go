package main

import (
	"fmt"
	"sync"
)

//Fetcher is an abstraction for Fetching content from urls
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

//FetchResult is a wrapper over the Fetch result
type FetchResult struct {
	body string
	urls []string
	err  error
}

//URL is an alias for readbility to a string of a url
type URL = string

//FetcherCache is a Cache to Fetch results faster, using the Proxy Pattern
type FetcherCache struct {
	//Delegator is the Fetcher that is being cached
	Delegator Fetcher
	//Cache mapping between a Url to a FetchResult
	Cache map[URL]*FetchResult
	lock  sync.Mutex
}

//Fetch is a implementation for FecherCache
func (f *FetcherCache) Fetch(url string) (body string, urls []string, err error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	fetchResult, isCached := f.Cache[url]
	if isCached {
		return fetchResult.body, fetchResult.urls, fetchResult.err
	}
	b, urls, err := f.Delegator.Fetch(url)
	f.Cache[url] = &FetchResult{
		body: b,
		urls: urls,
		err:  err,
	}
	return b, urls, err
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	for _, u := range urls {
		go Crawl(u, depth-1, fetcher)
	}
	return
}

func main() {
	Crawl("https://golang.org/", 4, &FetcherCache{
		Delegator: fetcher,
		Cache:     make(map[URL]*FetchResult),
	})
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
