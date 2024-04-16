// Create a new Go HTTP server
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", handleEmpty)
	router.HandleFunc("/ws", wsHandler).Methods("GET")
	router.HandleFunc("/confirm", handleConfirm).Methods("POST")
	router.HandleFunc("/request", handleRequest).Methods("POST")
	log.Println("Listening on 8080")
	http.ListenAndServe("localhost:8080", router)

}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("Client connected")
	conn.Close()
}

// Store the secret codes and corresponding socket IDs
var pendingRequests = make(map[string]map[string]interface{})
var socketIds = make(map[string]string)
var bindedSessions = make(map[string]bool)

func handleEmpty(w http.ResponseWriter, r *http.Request) {
	response := `{"Result": "Not Found"}`
	json.NewEncoder(w).Encode(response)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var reqBody struct {
		SecretCode string `json:"secretCode"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	secretCode := reqBody.SecretCode
	fmt.Println("User requested to connect with secretCode:", secretCode)

	pendingRequests[secretCode] = make(map[string]interface{})
	fmt.Println("Pending requests:", pendingRequests)

	// Check if the secret code is valid (e.g., 6 digits)
	if len(secretCode) != 6 {
		fmt.Println("Invalid secret code")
		http.Error(w, "Invalid secret code", http.StatusBadRequest)
		return
	}

	// Store the request in pendingRequests
	pendingRequests[secretCode]["receiver"] = r

	// Wait for the second party to hit the same URL
	// (You can implement a timeout here if needed)
	for i := 0; i < 60; i++ {
		senderPartyRequest := pendingRequests[secretCode]["sender"]
		if senderPartyRequest != nil {
			// The second party has confirmed the request
			senderPartySessionID := senderPartyRequest.(*http.Request).Header.Get("Session-Id")
			bindedSessions[senderPartySessionID+"-"+r.Header.Get("Session-Id")] = true
			fmt.Println("Both parties confirmed the request. Bind the sessions:", bindedSessions)

			// Clean up pendingRequests
			delete(pendingRequests, secretCode)

			// Respond with the session IDs
			response := struct {
				SessionID            string `json:"sessionId"`
				SenderPartySessionID string `json:"senderPartySessionId"`
			}{
				SessionID:            r.Header.Get("Session-Id"),
				SenderPartySessionID: senderPartySessionID,
			}
			json.NewEncoder(w).Encode(response)
			return
		} else {
			time.Sleep(1 * time.Second)
		}
	}

	// Respond with a success message
	http.Error(w, "Failed to connect to the device!", http.StatusBadRequest)
}

func handleConfirm(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var reqBody struct {
		SecretCode string `json:"secretCode"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	secretCode := reqBody.SecretCode
	fmt.Println("User confirmed to connect with secretCode:", secretCode)

	// Check if the secret code is valid (e.g., 6 digits)
	if len(secretCode) != 6 {
		http.Error(w, "Invalid secret code", http.StatusBadRequest)
		return
	}

	// Retrieve the first party's request from pendingRequests
	receiverPartyRequest := pendingRequests[secretCode]["receiver"]
	if receiverPartyRequest == nil {
		http.Error(w, "No pending request found", http.StatusNotFound)
		return
	}

	// Create a session for both parties
	sessionID := r.Header.Get("Session-Id")
	receiverPartySessionID := receiverPartyRequest.(*http.Request).Header.Get("Session-Id")

	// Store socket IDs for both parties
	socketIds[receiverPartySessionID] = ""
	socketIds[sessionID] = ""

	// Respond with the session IDs
	response := struct {
		SessionID              string `json:"sessionId"`
		ReceiverPartySessionID string `json:"receiverPartySessionId"`
	}{
		SessionID:              sessionID,
		ReceiverPartySessionID: receiverPartySessionID,
	}
	json.NewEncoder(w).Encode(response)

	// Clean up pendingRequests (optional)
	delete(pendingRequests, secretCode)
}
