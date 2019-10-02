package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
    "net/http"
    "os"
    "strconv"

    "github.com/gorilla/mux"
    "github.com/google/uuid"
)

type DiceThrowResponsePayload struct {
	Faces int
	Eyes  int
	User  string
}

type GetAuthCodeURLResponsePayload struct {
	AuthorizeUrl string
}

type BadRequestPayload struct {
	Reason string
}

var client_id = os.Getenv("CLIENT_ID")
var client_secret = os.Getenv("CLIENT_SECRET")
var github_authorize_url = "https://github.com/login/oauth/authorize"
var github_access_token_url = "https://github.com/login/oauth/access_token"
var github_user_api_url = "https://api.github.com/user"

func ThrowDiceHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	access_token := r.URL.Query().Get("access_token")
	userName := "anonymous"
	if (access_token != "") {
		client := &http.Client{}
		githubRequest, err := http.NewRequest("GET", github_user_api_url + "?access_token=" + access_token, nil)
		githubRequest.Header.Add("Accept", "application/json")
		githubRequest.Header.Add("User-Agent", "dice-test")
		resp, err := client.Do(githubRequest)
		if (err != nil) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if (err != nil) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var responsePayload map[string]interface{}
		json.Unmarshal([]byte(body), &responsePayload)
		userName = responsePayload["name"].(string)
	}

	faceCount, error := strconv.Atoi(params["faceCount"])
	if (error != nil) {
		http.Error(w, error.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode(BadRequestPayload{Reason: error.Error()})
		return
	}
	if (faceCount <= 0) {
		http.Error(w, error.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode(BadRequestPayload{Reason: "Number cannot be negative."})
		return
	}
	eyesCount := rand.Intn(faceCount) + 1
	responsePayload := DiceThrowResponsePayload{Faces: faceCount, Eyes: eyesCount, User: userName}

	json.NewEncoder(w).Encode(responsePayload)
	w.WriteHeader(http.StatusOK)
}

func GetAuthCodeURLHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	randomUUID, err := uuid.NewRandom()
	if (err != nil) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	state := uuid.UUID.String(randomUUID)
	redirectURL := github_authorize_url + "?state=" + state + "&client_id=" + client_id
	responsePayload := GetAuthCodeURLResponsePayload{AuthorizeUrl: redirectURL}
	json.NewEncoder(w).Encode(responsePayload)
	w.WriteHeader(http.StatusOK)
}

func GetAccessTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	requestBody, err := json.Marshal(map[string]string{
		"client_id": client_id,
		"state": r.URL.Query().Get("state"),
		"code": r.URL.Query().Get("code"),
		"client_secret": client_secret,
	})

	if (err != nil) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	authRequest, err := http.NewRequest("POST", github_access_token_url, bytes.NewBuffer(requestBody))
	authRequest.Header.Add("Accept", "application/json")
	authRequest.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	authResponse, err := client.Do(authRequest)
	if (err != nil) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer authResponse.Body.Close()

	body, err := ioutil.ReadAll(authResponse.Body)
	if (err != nil) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var responsePayload map[string]interface{}
	json.Unmarshal([]byte(body), &responsePayload)

	http.Redirect(w, r, "http://users.metropolia.fi/~adarshk/?access_token=" +
				  responsePayload["access_token"].(string), 302)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/diceapi/dice/{faceCount:[0-9]+}", ThrowDiceHandler)
	router.HandleFunc("/diceapi/authorization/get_code", GetAuthCodeURLHandler)
	router.HandleFunc("/diceapi/authorization/get_token", GetAccessTokenHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}
