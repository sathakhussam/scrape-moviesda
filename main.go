package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
)

type Movie struct {
	Name string `json:"name"`
	Url string `json:"url"`
}

func SaveImage(i int, s *goquery.Selection) {
	url, yes := s.Attr("src")
	if !yes {
		url = ""
	} else {
		url = "https://moviesda9.com" + url
	}
	rqImage, err := http.Get(url)
	for err != nil {
		rqImage, err = http.Get(url)
	}
	defer rqImage.Body.Close()
	st, err := io.ReadAll(rqImage.Body)
	if err != nil {
		log.Fatal(err)
	}
	id := uuid.New()
	os.WriteFile("images/"+id.String()+".jpg", st, 0655)
}

func main() {
	// GetOverview(20)
	tm := time.Now()
	allMoviesRd, err := os.ReadFile("data-overview.json")
	if err != nil {
		log.Println(err)
	}
	allMovies := []Movie{}
	err = json.Unmarshal(allMoviesRd, &allMovies)
	if err != nil {
		log.Println(err)
	}

	var wg sync.WaitGroup

	for i, movie := range allMovies {
		wg.Add(1)
		go GetMovie(movie, &wg)

		// Limit the number of concurrent goroutines
		if (i+1)%512 == 0 {
			wg.Wait() // Wait for the current batch to finish
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println(time.Since(tm))
}

func GetMovie(movie Movie, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println(movie.Name)
	req, err := http.Get(movie.Url)
	for err != nil {
		req, err = http.Get(movie.Url)
	}
	defer req.Body.Close()
	rd, err := goquery.NewDocumentFromReader(req.Body)
	if err != nil {
		log.Println(err)
	}
	rd.Find("img.tbl").Each(SaveImage)
	rd.Find("div.mvscreen img").Each(SaveImage)
}

func GetOverview(maxPages int) {
	movies := []Movie{}
	for i := 1; i < maxPages+1; i++ {
		req, err := http.Get("https://moviesda9.com/tamil-2023-movies/?page=" + strconv.Itoa(i))
		for err != nil {
			req, err = http.Get("https://moviesda9.com/tamil-2023-movies/?page=" + strconv.Itoa(i))
		}
		defer req.Body.Close()
		rd, err := goquery.NewDocumentFromReader(req.Body)
		if err != nil {
			log.Println(err)
		}
		rd.Find("div.f a").Each(func(i int, s *goquery.Selection) {
			url, yes := s.Attr("href")
			if !yes {
				url = ""
			} else {
				url = "https://moviesda9.com" + url
			}
			movies = append(movies, Movie{
				Name: s.Text(),
				Url: url,
			})
		})
		fmt.Println(movies)
	}
	jr, err := json.Marshal(movies)
	if err != nil {
		log.Println(err)
	}
	os.WriteFile("data-overview.json", jr, 0655)
}