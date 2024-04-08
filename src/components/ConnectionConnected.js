import React, { useState } from 'react';

const ConnectionConnected = ({ connectionState, setConnectionState, handleDisconnect, handleSendMessage }) => {
    const [textToSend, setTextToSend] = useState('');
    const handleSend = () => {
        // Logic to send the message
        handleSendMessage(textToSend);
    };

    return (
        <div>
            <p>Connected to: {connectionState.to}</p>
            <input type="text" placeholder="Text to send" value={textToSend} onChange={(e) => setTextToSend(e.target.value)}/>
            <button onClick={handleSend}>Send</button>
            <button onClick={handleDisconnect}>Disconnect</button>
        </div>
    );
};

export default ConnectionConnected;