// import { io } from "socket.io-client";
const {io} = require('socket.io-client');
console.log("Starting connection..");
const socket = io("http://localhost:3000", {
    withCredentials: true,
}); // Assuming you have the socket.io library included

socket.on("secretCode", (data) => {
    console.log(data);
    console.log("Connected to server");
});

socket.on("disconnect", () => {
    console.log("Disconnected from server");
});