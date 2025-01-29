package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleStatus(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedResponse string
	}{
		{"https://www.google.com", http.StatusOK,"UP"},
		{"https://www.facebook.com", http.StatusOK,"UP"},
		{"https://www.x.com", http.StatusOK,"UP"},
		{"https://www.abnfr.com", http.StatusBadRequest,"INVALID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/websites?name="+tt.name, nil)
			rec := httptest.NewRecorder()
			
			HandleStatus(rec, req)

			var response map[string]string
			err:=json.Unmarshal(rec.Body.Bytes(), &response)
			if err!=nil{
				t.Fatalf("Failed to parse response %v", err)
			}
			
			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d for %v, got %d", tt.expectedStatus, tt.name, rec.Code)
			
			}
	
			if response[tt.name]!=tt.expectedResponse{
				t.Errorf("Expected response %s for %v, got %s",tt.expectedResponse, tt.name, response[tt.name])
			}

		})
	}
}


func TestPostWebsitesList( t *testing.T){
	tests:=[]struct{
		name string
		input string 
		expectedResponse string

	}{
		{"Empty Body",``,"No websites added"},
		{"Valid websites list",`{"data":["https://www.google.com","https://www.x.com"]}`,"Websites added successfully"},
	}

	for _,tt:=range tests{
		req:=httptest.NewRequest("POST","/websites",bytes.NewBuffer([]byte(tt.input)))
		rec:=httptest.NewRecorder()

		PostWebsitesList(rec,req)

		var response string
		err:=json.Unmarshal(rec.Body.Bytes(),&response)

		if err!=nil{
				t.Fatalf("Error sending request %v",err)
		}

		if response!=tt.expectedResponse{
			t.Errorf("Expected response %s for %v, Actual response %s",tt.expectedResponse,tt.name,response)
		}
	}
}


