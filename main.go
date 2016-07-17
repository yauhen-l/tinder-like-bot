package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/yauhen-l/tinder"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
)

var t *tinder.Tinder
var likes = 0
var matches = 0
var yesLimit int
var keepSearching = true
var filter Filter
var dryRun bool

type Filter struct {
	ExcludeName     []string
	Schools         []string
	CommonInterests []string
}

func main() {
	fbtoken := flag.String("fb-token", "", "facebook access_token")
	flag.BoolVar(&dryRun, "dry-run", false, "do not like or pass")
	flag.IntVar(&yesLimit, "yes-limit", 30, "limit of likes to stop")
	filterFile := flag.String("filter", "filter.json", "path to filter file")
	flag.Parse()

	filterJson, err := ioutil.ReadFile(*filterFile)
	if err != nil {
		log.Fatalf("failed to read filter file %q due: %s", *filterFile, err)
	}
	err = json.Unmarshal(filterJson, &filter)
	if err != nil {
		log.Fatalf("failed to parse filter file %s due: %s", *filterFile, err)
	}

	fbUserId, err := getFacebookUserId(*fbtoken)
	if err != nil {
		log.Fatalf("failed to get facebook user ID: %s", err)
	}

	t = tinder.Init(fbUserId, *fbtoken)
	err = t.Auth()
	if err != nil || len(t.Me.Token) == 0 {
		log.Fatalf("failed to authenticate into Tinder: %s", err)
	}

	log.Printf("Logged in into Tinder as: %+v", t.Me)

	likeIdCh := make(chan string)

	go findRecommendations(likeIdCh)

	for id := range likeIdCh {
		match, err := like(id)
		if err != nil {
			log.Printf("failed to like %s due: %s", id, err)
			continue
		}

		log.Printf("you like %s. match=%v", id, match)
		likes++
		if match {
			matches++
		}
		if likes >= yesLimit {
			keepSearching = false
			log.Printf("already liked %d", likes)
			break
		}
	}
	log.Printf("liked this time: %d, macthed: %d", likes, matches)
	log.Println("exit")
}

func like(id string) (bool, error) {
	if dryRun {
		return false, nil
	}
	return t.Like(id)
}

func pass(id string) error {
	if dryRun {
		return nil
	}
	return t.Pass(id)
}

func findRecommendations(likeIdCh chan<- string) {
	defer close(likeIdCh)

	for keepSearching {
		resp, err := t.GetRecommendations(yesLimit)
		if err != nil {
			log.Printf("failed to get recommendations: %s", err)
			return
		}

		wg := sync.WaitGroup{}

		for _, rec := range resp.Results {
			log.Printf("recommendations: %+v", rec.Name)
			wg.Add(1)

			go func(r tinder.Recommendation) {
				defer wg.Done()

				if result, ok := match(r); ok {
					log.Printf("like id %s, name=%s, because: %s", r.ID, r.Name, result)
					likeIdCh <- r.ID
				} else {
					log.Printf("pass id %s, name=%s, because: %s", r.ID, r.Name, result)
					pass(r.ID)
				}
			}(rec)
		}

		wg.Wait()
	}
}

func containsAny(text []string, keys []string) (string, bool) {
	for _, part := range text {
		for _, key := range keys {
			if strings.Contains(strings.ToLower(part), strings.ToLower(key)) {
				return key, true
			}
		}
	}

	return "", false
}

func match(rec tinder.Recommendation) (string, bool) {
	if name, ok := containsAny([]string{rec.Name}, filter.ExcludeName); ok {
		return fmt.Sprintf("name exlusion: %s", name), false
	}

	if school, ok := containsAny(rec.SchoolsNames(), filter.Schools); ok {
		return fmt.Sprintf("school matched: %s", school), true
	}

	if intereset, ok := containsAny(append(rec.CommonInterestsNames(), rec.Bio), filter.CommonInterests); ok {
		return fmt.Sprintf("interest matched: %s", intereset), true
	}

	return "nothing matched", false
}

func getFacebookUserId(fbtoken string) (string, error) {
	if len(fbtoken) == 0 {
		return "", errors.New("fb_token is required")
	}

	resp, err := http.Get("https://graph.facebook.com/me?access_token=" + fbtoken)
	if err != nil {
		return "", fmt.Errorf("failed to access facebook: %s", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read facebook response: %s", err)
	}

	fbMe := struct {
		Name  string
		Id    string
		Error interface{}
	}{}

	err = json.Unmarshal(data, &fbMe)
	if err != nil {
		return "", fmt.Errorf("failed to parse facebook response: %s", err)
	}

	if fbMe.Error != nil {
		return "", fmt.Errorf("facebook returned error: %v", fbMe.Error)
	}

	return fbMe.Id, nil
}
