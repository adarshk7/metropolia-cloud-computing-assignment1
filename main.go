package main

import (
	"encoding/json"
	"log"
	"math/rand"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
)

type DiceThrowResponsePayload struct {
	Faces int
	Eyes  int
}

type BadRequestPayload struct {
	Reason string
}

func ThrowDiceHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

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
	eyesCount := rand.Intn(faceCount - 1) + 1
	responsePayload := DiceThrowResponsePayload{Faces: faceCount, Eyes: eyesCount}

	json.NewEncoder(w).Encode(responsePayload)
	w.WriteHeader(http.StatusOK)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/diceapi/dice/{faceCount:[0-9]+}", ThrowDiceHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}
