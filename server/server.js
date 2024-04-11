
// Create an Express app
const express = require('express');
const app = express();
const sessions = require('express-sessions');
const http = require('http');
const httpServer = http.createServer(app);
const { Server } = require("socket.io");
const io = new Server(httpServer, {
    cors: {
        origin: ["http://localhost:3001"],
        credentials: true
    }
});

// Store the secret codes and corresponding socket IDs
const pendingRequests = {};
const socketIds = {};
const bindedSessions = {};

// Generate a random 6-digit alphanumeric secret code
function sleepFor(sleepDuration){
    var now = new Date().getTime();
    while(new Date().getTime() < now + sleepDuration){ /* Do nothing */ }
}

// Route to handle the initial request with a secret code
app.post('/request', (req, res) => {
    const { secretCode } = req.body;

    pendingRequests[secretCode] = {
        "receiver": null,
        "sender": null,
    };

    // Check if the secret code is valid (e.g., 6 digits)
    if (!/^\d{6}$/.test(secretCode)) {
        return res.status(400).json({ error: 'Invalid secret code' });
    }

    // Store the request in pendingRequests
    pendingRequests[secretCode]["receiver"] = req;

    // Wait for the second party to hit the same URL
    // (You can implement a timeout here if needed)
    for(let i = 0; i < 60; i++) {
        const senderPartyRequest = pendingRequests[secretCode]["sender"];
        if ( senderPartyRequest ) {
            // The second party has confirmed the request
            const senderPartySessionId = senderPartyRequest.session.id;
            bindedSessions[senderPartySessionId + "-" + req.session.id] = true;

            // Clean up pendingRequests
            delete pendingRequests[secretCode];

            // Respond with the session IDs
            res.json({ sessionId: req.session.id, senderPartySessionId: senderPartySessionId });
            return;
        } else {
            i++;
            sleepFor(1000)
        }
    }

    // Respond with a success message
    res.status(400).json({ message: 'Failed to connect to the device!' });
});

// Route to handle the second party's request with the same secret code
app.post('/confirm', (req, res) => {
    const { secretCode } = req.body;

    // Check if the secret code is valid (e.g., 6 digits)
    if (!/^\d{6}$/.test(secretCode)) {
        return res.status(400).json({ error: 'Invalid secret code' });
    }

    // Retrieve the first party's request from pendingRequests
    const receiverPartyRequest = pendingRequests[secretCode]["receiver"];

    if (!receiverPartyRequest) {
        return res.status(404).json({ error: 'No pending request found' });
    }

    // Create a session for both parties
    const sessionId = req.session.id;
    const receiverPartySessionId = receiverPartyRequest.session.id;

    // Store socket IDs for both parties
    socketIds[receiverPartySessionId] = null;
    socketIds[sessionId] = null;

    // Respond with the session IDs
    res.json({ sessionId, receiverPartySessionId });

    // Clean up pendingRequests (optional)
    delete pendingRequests[secretCode];
});

// Socket.IO connection
io.on('connection', (socket) => {
    const sessionId = socket.request.session.id;

    // Store the socket ID for the connected party
    socketIds[sessionId] = socket.id;

    // Listen for messages from the first party
    socket.on('message', (data) => {
        const { toSessionId, message } = data;

        // Get the socket ID of the recipient
        const recipientSocketId = socketIds[toSessionId];

        if (recipientSocketId) {
            if ( sessionId + "-" + toSessionId in bindedSessions && bindedSessions[sessionId + "-" + toSessionId]) {
                // Send the message to the recipient
                io.to(recipientSocketId).emit('message', { fromSessionId: sessionId, message });
            } else {
                console.log(`Session is not binded.`);
            }
        } else {
            // Handle recipient not found (optional)
            console.log(`Recipient with session ID ${toSessionId} not found.`);
        }
    });

    socket.on("disconnect", (data) => {
        console.log(`Socket ${socketIds[sessionId]} disconnected.`);
        delete socketIds[sessionId];
    })
});

httpServer.listen(3000, () => {
    console.log('listening on *:3000');
});
