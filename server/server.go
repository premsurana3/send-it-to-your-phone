// Create a new Go HTTP server
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/googollee/go-socket.io"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
)

// Store the secret codes and corresponding socket IDs
var pendingRequests = make(map[string]map[string]interface{})
var socketIds = make(map[string]string)
var bindedSessions = make(map[string]bool)

// Generate a random 6-digit alphanumeric secret code
func generateSecretCode() string {
    // Generate a random 6-digit alphanumeric code
    // implementation here...
    return "ABC123"
}

// Route to handle the initial request with a secret code
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
                SessionID           string `json:"sessionId"`
                SenderPartySessionID string `json:"senderPartySessionId"`
            }{
                SessionID:           r.Header.Get("Session-Id"),
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

// Route to handle the second party's request with the same secret code
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
        SessionID           string `json:"sessionId"`
        ReceiverPartySessionID string `json:"receiverPartySessionId"`
    }{
        SessionID:           sessionID,
        ReceiverPartySessionID: receiverPartySessionID,
    }
    json.NewEncoder(w).Encode(response)

    // Clean up pendingRequests (optional)
    delete(pendingRequests, secretCode)
}

// Socket.IO connection
func handleSocketConnection(so socketio.Socket) {
    sessionID := so.Request().Header.Get("Session-Id")

    // Store the socket ID for the connected party
    socketIds[sessionID] = so.Id()

    // Listen for messages from the first party
    so.On("message", func(data string) {
        // Parse the message data
        var messageData struct {
            ToSessionID string `json:"toSessionId"`
            Message     string `json:"message"`
        }
        err := json.Unmarshal([]byte(data), &messageData)
        if err != nil {
            log.Println("Error parsing message data:", err)
            return
        }

        toSessionID := messageData.ToSessionID
        message := messageData.Message

        // Get the socket ID of the recipient
        recipientSocketID := socketIds[toSessionID]

        if recipientSocketID != "" {
            if bindedSessions[sessionID+"-"+toSessionID] {
                // Send the message to the recipient
                so.BroadcastTo(recipientSocketID, "message", map[string]interface{}{
                    "fromSessionId": sessionID,
                    "message":       message,
                })
            } else {
                fmt.Println("Session is not binded.")
            }
        } else {
            // Handle recipient not found (optional)
            fmt.Printf("Recipient with session ID %s not found.\n", toSessionID)
        }
    })

    so.On("disconnection", func() {
        fmt.Printf("Socket %s disconnected.\n", socketIds[sessionID])
        delete(socketIds, sessionID)
    })
}

func main() {
    // Create a new router
    r := mux.NewRouter()

    // Create a new session store
    store := sessions.NewCookieStore([]byte("your-secret-key"))

    // Route to handle the initial request with a secret code
    r.HandleFunc("/request", handleRequest).Methods("POST")

    // Route to handle the second party's request with the same secret code
    r.HandleFunc("/confirm", handleConfirm).Methods("POST")

    // Serve static files
    r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

    // Create a new Socket.IO server
    server, err := socketio.NewServer(nil)
    if err != nil {
        log.Fatal(err)
    }

    // Handle Socket.IO connections
    server.On("connection", handleSocketConnection)

    // Attach the Socket.IO server to the router
    r.Handle("/socket.io/", server)

    // Start the HTTP server
    log.Println("Listening on :3000...")
    log.Fatal(http.ListenAndServe(":3000", r))
}