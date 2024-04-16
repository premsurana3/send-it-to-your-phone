// Create a new Go HTTP server
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", handleEmpty)
	router.HandleFunc("/ws", handler).Methods("GET")
	router.HandleFunc("/confirm", handleConfirm).Methods("POST")
	router.HandleFunc("/request", handleRequest).Methods("POST")
	log.Println("Listening on 8080")
	http.ListenAndServe("localhost:8080", router)

}

func handler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-id")
	sessionId := session.ID

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer conn.Close()

	socketIds[sessionId] = conn

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		toSessionId := string(message)
		recipientSocketId, ok := socketIds[toSessionId]

		if ok {
			if bindedSessions[sessionId+"-"+toSessionId] {
				err = recipientSocketId.WriteMessage(websocket.TextMessage, []byte(sessionId+": "+string(message)))
				if err != nil {
					log.Println("write:", err)
					break
				}
			} else {
				log.Println("Session is not binded.")
			}
		} else {
			log.Println("Recipient with session ID", toSessionId, "not found.")
		}
	}

	log.Println("Socket", sessionId, "disconnected.")
	delete(socketIds, sessionId)
}

// Store the secret codes and corresponding socket IDs

var pendingRequests = make(map[string]map[string]*http.Request)
var socketIds = make(map[string]*websocket.Conn)
var bindedSessions = make(map[string]bool)

func handleEmpty(w http.ResponseWriter, r *http.Request) {
	response := `{"Result": "Not Found"}`
	json.NewEncoder(w).Encode(response)
}

func cleanUpPendingRequest(secretCode string) {
	// delete from store if available in future
	delete(pendingRequests, secretCode)
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
	log.Println("User requested to connect with secretCode:", secretCode)

	pendingRequests[secretCode] = make(map[string]*http.Request)
	log.Println("Pending requests:", pendingRequests)

	// Check if the secret code is valid (e.g., 6 digits)
	if len(secretCode) != 6 {
		log.Println("Invalid secret code")
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
			senderPartySession, err := store.Get(senderPartyRequest, "session")

			if err != nil {
				cleanUpPendingRequest(secretCode)
				http.Error(w, "No sender request found", http.StatusNotFound)
				return
			}
			senderPartySessionID := senderPartySession.ID
			session, _ := store.Get(r, "session")
			session.Save(r, w)

			bindedSessions[senderPartySessionID+"-"+session.ID] = true
			log.Println("Both parties confirmed the request. Bind the sessions: ", senderPartySessionID+"-"+session.ID)

			// Clean up pendingRequests
			cleanUpPendingRequest(secretCode)

			// Respond with the session IDs
			response := struct {
				SessionID            string `json:"sessionId"`
				SenderPartySessionID string `json:"senderPartySessionId"`
			}{
				SessionID:            session.ID,
				SenderPartySessionID: senderPartySessionID,
			}

			json.NewEncoder(w).Encode(response)
			return
		} else {
			log.Println("Waiting... ", i, " seconds passed")
			time.Sleep(1 * time.Second)
		}
	}

	cleanUpPendingRequest(secretCode)

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
	log.Println("User confirmed to connect with secretCode:", secretCode)

	// Check if the secret code is valid (e.g., 6 digits)
	if len(secretCode) != 6 {
		log.Println("Invalid secret code")
		http.Error(w, "Invalid secret code", http.StatusBadRequest)
		cleanUpPendingRequest(secretCode)
		return
	}

	// Retrieve the first party's request from pendingRequests
	receiverPartyRequest := pendingRequests[secretCode]["receiver"]
	if receiverPartyRequest == nil {
		log.Println("No pending request found")
		http.Error(w, "No pending request found", http.StatusNotFound)
		cleanUpPendingRequest(secretCode)
		return
	}

	// Create a session for both parties
	session, _ := store.Get(r, "session")
	session.Save(r, w)

	pendingRequests[secretCode]["sender"] = r
	sessionID := session.ID

	receiverSession, err := store.Get(receiverPartyRequest, "session")
	if err != nil {
		http.Error(w, "No receiver session found", http.StatusNotFound)
		return
	}
	receiverPartySessionID := receiverSession.ID

	// Store socket IDs for both parties
	socketIds[receiverPartySessionID] = nil
	socketIds[sessionID] = nil

	// Respond with the session IDs
	response := struct {
		SessionID              string `json:"sessionId"`
		ReceiverPartySessionID string `json:"receiverPartySessionId"`
	}{
		SessionID:              sessionID,
		ReceiverPartySessionID: receiverPartySessionID,
	}
	log.Println(sessionID)
	log.Println(receiverPartySessionID)
	log.Printf("User confirmed: %v", pendingRequests[secretCode])
	json.NewEncoder(w).Encode(response)
}
