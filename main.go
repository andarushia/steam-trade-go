package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Philipp15b/go-steamapi"
)

var promptData string
var templates *template.Template

const appId uint64 = 753
const key string = "53487FD6BD8A52B4980A3E099BD5A435"
const apiUrl string = "http://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/?key="

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
	if r.Method == http.MethodGet {
		templates.Execute(w, promptData)
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		promptData = r.FormValue("data")
		steamId, err := convertToSteamID(promptData)
		if err != nil {
			// Handle the error, e.g., display an error message on the webpage.
			http.Error(w, "Invalid Steam username or ID", http.StatusBadRequest)
			return
		}

		inv, err := steamapi.GetPlayerItems(steamId, appId, key)

		if err != nil {
			http.Error(w, "Inventory is not accessible", http.StatusBadRequest)
		}

		fmt.Println(inv)

		if err := templates.ExecuteTemplate(w, "index.html", steamId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func convertToSteamID(input string) (uint64, error) {
	// Check if the input is already a numeric Steam ID (e.g., "76561198012345678").
	if strings.HasPrefix(input, "7656119") && len(input) == 17 {
		out, err := stringToUint64(input)
		if err != nil {
			return 0, err
		}
		return out, nil
	}

	// If it's not a numeric ID, it might be a vanity URL.
	input = strings.TrimPrefix(input, "https://steamcommunity.com/id/")
	input = strings.TrimSuffix(input, "/")
	idRequest := apiUrl + key + "&vanityurl=" + input
	request, err := http.NewRequest("GET", idRequest, nil)

	if err != nil {
		return 0, err
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		return 0, err
	}

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	steamId, success := formatJson(responseBody)

	if success == 42 {
		return 0, errors.New("steam responded with failure getting user's id")
	}

	fmt.Println(steamId)
	return steamId, nil
}

func formatJson(input []byte) (uint64, uint64) {
	out := string(input)
	out = strings.TrimPrefix(out, "{\"response\":{\"steamid\":\"")
	steamId, err := stringToUint64(out[:16])
	if err != nil {
		fmt.Println(err)
	}
	out = strings.TrimPrefix(out[17:], "\",\"success\":")
	success, err := stringToUint64(strings.TrimSuffix(out, "}}"))
	if err != nil {
		fmt.Println(err)
	}
	return steamId, success
}

func stringToUint64(input string) (uint64, error) {
	out, err := strconv.Atoi(input)
	if err != nil {
		return 0, err
	}
	return uint64(out), nil
}
