package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
	// "errors"
)

type StatusChecker interface {
	Check(ctx context.Context, name string) (status bool, err error)
}

type httpChecker struct {
}

func (h httpChecker) Check(ctx context.Context, name string) (status bool, err error) {
	response, err := http.Get(name)
	if err != nil || response.StatusCode != 200 {
		return false, err
	}
	return true, nil

}

type websites struct {
	Data map[string]bool `json:"data"`
}

var res = &websites{
	Data: make(map[string]bool),
}

var websiteStatus = map[string]string{}

func main() {
	var mutex sync.Mutex
	go func() {
		for {
			mutex.Lock()
			defer mutex.Unlock()
			updateStatus()
			time.Sleep(1 * time.Minute)

		}

	}()

	mux := http.DefaultServeMux

	mux.HandleFunc("POST /websites", PostWebsitesList)
	mux.HandleFunc("GET /websites", HandleStatus)

	fmt.Print("Server running on port 3000")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		fmt.Println(err)
	}
}

func HandleStatus(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()
	val := params.Get("name")

	if val != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		var a httpChecker
		status, err := a.Check(ctx, val)
		var responseCode = http.StatusOK
		if !status {
			websiteStatus[val] = "DOWN"
		} else {
			websiteStatus[val] = "UP"
		}
		if err != nil {
			websiteStatus[val] = "INVALID"
			responseCode = http.StatusBadRequest

		}

		handleResponse(w, map[string]string{val: websiteStatus[val]}, responseCode)
	} else {
		updateStatus()
		handleResponse(w, websiteStatus, http.StatusOK)
	}
}

func PostWebsitesList(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	websitesList := &struct {
		Data []string `json:"data"`
	}{Data: nil}

	err = json.Unmarshal(body, &websitesList)
	if websitesList.Data == nil {
		handleResponse(w, "No websites added", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
		return
	}

	for _, val := range websitesList.Data {
		if _, ok := res.Data[val]; !ok {
			res.Data[val] = false
		}
	}
	fmt.Println("i m here")
	handleResponse(w, "Websites added successfully", http.StatusOK)
}

func updateStatus() {
	for url := range res.Data {

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
		defer cancel()
		var a httpChecker
		status, err := a.Check(ctx, url)
		if err != nil {
			res.Data[url] = false
			websiteStatus[url] = "DOWN"
			fmt.Println("Error getting response from website", err)
			continue
		}
		if status {
			res.Data[url] = true
			websiteStatus[url] = "UP"

		} else {
			res.Data[url] = false
			websiteStatus[url] = "DOWN"
		}
	}

}

func handleResponse(w http.ResponseWriter, message any, statusCode int) {

	response, err := json.Marshal(message)
	if err != nil {
		http.Error(w, "Error converting response to json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}
