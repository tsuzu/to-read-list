package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v48/github"
	"github.com/tsuzu/to-read-list/pkg/issue"
	"github.com/tsuzu/to-read-list/pkg/summarizer"
)

func main() {
	appID, err := strconv.ParseInt(os.Getenv("APP_ID"), 10, 64)

	if err != nil {
		panic(err)
	}

	installationID, err := strconv.ParseInt(os.Getenv("APP_INSTALLATION_ID"), 10, 64)

	if err != nil {
		panic(err)
	}

	tr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, []byte(os.Getenv("APP_PRIVATE_KEY")))

	if err != nil {
		panic(err)
	}

	client := github.NewClient(&http.Client{
		Transport: tr,
	})

	type Body struct {
		URL string `json:"url"`
	}

	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")

	http.ListenAndServe(":"+os.Getenv("PORT"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		var body Body
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			log.Println(err)

			return
		}

		meta, err := summarizer.GetMetadata(body.URL)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			log.Println(err)

			return
		}

		url, err := issue.Create(r.Context(), client, owner, repo, meta)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			log.Println(err)

			return
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Body{
			URL: url,
		})
	}))
}
