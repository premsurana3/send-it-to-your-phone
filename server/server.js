
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
const secretCodes = {};
const connections = {};


// Generate a random 6-digit alphanumeric secret code
function generateSecretCode() {
    const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    let code = '';
    for (let i = 0; i < 6; i++) {
        code += characters.charAt(Math.floor(Math.random() * characters.length));
    }
    return code;
}

// Handle socket connections
io.on('connection', (socket) => {
    console.log('A user connected');
    console.log(socket.handshake.query.type)
    // Generate a secret code for the connected socket
    
    if ( socket.handshake.query.type === 'receiver' ) {
        let secretCode = generateSecretCode();
        console.log("SecretCode: ", secretCode)
        secretCodes[secretCode] = {
            "receiver": socket.id,
            "sender": null,
        };
        io.to(socket.id).emit('secretCode', secretCode);
    } else if ( socket.handshake.query.type === 'sender' ) {
        let secretCode = socket.handshake.query.code;
        try{
            secretCodes[secretCode].sender = socket.id;
        }catch(e){
            console.log('Invalid secret code');
            socket.emit('error', 'Invalid secret code');
            socket.disconnect(true);
            return;
        }
    }
    
    // Listen for incoming messages
    socket.on('message', (data) => {
        const { code, message } = data;

        // Check if the secret code is valid
        if (secretCodes[code]) {
            // Get the socket ID of the recipient
            if ( secretCodes[code].sender === socket.id ) {
                let recipientSocketId = secretCodes[code]['receiver'];
                // Send the message to the recipient
                io.to(recipientSocketId).emit('message', message);
            }
        } else {
            // Invalid secret code
            socket.emit('error', 'Invalid secret code');
        }
    });

    // Handle socket disconnections
    socket.on('disconnect', () => {
        console.log('A user disconnected');
            
        // Find the secret code associated with the disconnected socket
        const code = Object.keys(secretCodes).find((key) => {
            const { receiver, sender } = secretCodes[key];
            return receiver === socket.id || sender === socket.id;
        });

        if (code) {
            // Disconnect both sender and receiver
            const { receiver, sender } = secretCodes[code];

            io.sockets.sockets.forEach((socket) => {
                // If given socket id is exist in list of all sockets, kill it
                if( socket.id === receiver || socket.id === sender )
                    socket.disconnect(true);
            });

            // Remove the secret code from the storage
            delete secretCodes[code];
        }
    });
});

httpServer.listen(3000, () => {
    console.log('listening on *:3000');
});
