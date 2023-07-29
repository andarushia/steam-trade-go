package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var promptData string
var templates *template.Template

const (
	appId     uint64 = 753
	contendId uint64 = 6
	key       string = "53487FD6BD8A52B4980A3E099BD5A435"
	apiUrl    string = "http://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/?key=%s&vanityurl=%s"
)

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

		inv, err := getPlayerItems(steamId, appId, contendId)

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
		out, err := strconv.ParseUint(input, 10, 0)
		if err != nil {
			return 0, err
		}
		return out, nil
	}

	// If it's not a numeric ID, it might be a vanity URL.
	input = strings.TrimPrefix(input, "https://steamcommunity.com/id/")
	input = strings.TrimSuffix(input, "/")

	idRequest := fmt.Sprintf(apiUrl, key, input)
	jason, err := getJson(idRequest)
	if err != nil {
		return 0, err
	}

	steamId, success := parseId(jason)

	if success == 42 {
		return 0, errors.New("steam responded with failure while getting user's id")
	}

	return steamId, nil
}

func parseId(input []byte) (uint64, uint64) {
	out := string(input)
	out = strings.TrimPrefix(out, "{\"response\":{\"steamid\":\"")
	steamId, err := strconv.ParseUint(out[:17], 10, 0)
	if err != nil {
		fmt.Println(err)
	}
	out = strings.TrimPrefix(out[17:], "\",\"success\":")
	success, err := strconv.ParseUint(strings.TrimSuffix(out, "}}"), 10, 0)
	if err != nil {
		fmt.Println(err)
	}
	return steamId, success
}

func getPlayerItems(steamId uint64, appId uint64, contentId uint64) (string, error) {
	url := fmt.Sprintf("https://steamcommunity.com/inventory/%d/%d/%d", steamId, appId, contentId)
	inv, err := getJson(url)
	fmt.Println(url)

	if err != nil {
		return "", nil
	}

	return formatJson(inv), nil
}

func getJson(url string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		return nil, err
	}

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return responseBody, nil
}

func formatJson(data []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", " ")

	if err != nil {
		fmt.Println(err)
	}

	d := out.Bytes()
	return string(d)
}
