package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/Philipp15b/go-steamapi"
)

var promptData string
var templates *template.Template

const key string = "142412412155152"

func main() {
	templates = template.Must(template.ParseFiles("templates/index.html"))
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", homePage)
	fmt.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	var steamID string
	var err error
	if r.Method == http.MethodGet {
		templates.Execute(w, promptData)
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		promptData = r.FormValue("data")
		promptData, err = convertToSteamID(promptData)
		if err != nil {
			// Handle the error, e.g., display an error message on the webpage.
			http.Error(w, "Invalid Steam username or ID", http.StatusBadRequest)
			return
		}

		if err := templates.ExecuteTemplate(w, "index.html", steamID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func convertToSteamID(input string) (string, error) {
	// Check if the input is already a numeric Steam ID (e.g., "76561198012345678").
	if strings.HasPrefix(input, "7656119") && len(input) == 17 {
		return input, nil
	}

	// If it's not a numeric ID, it might be a vanity URL.
	input = strings.TrimPrefix(input, "https://steamcommunity.com/id/")
	input = strings.TrimSuffix(input, "/")

	steamID, err := steamapi.NewIdFromVanityUrl(input, key)
	if err != nil {
		return "", err
	}

	return steamID.String(), nil
}
