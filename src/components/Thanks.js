import React from 'react';

const Thanks = ({ connectionState, setConnectionState }) => {
    const handleClick = () => {
        setConnectionState({ status: 'disconnected' });
    };

    return (
        <div>
            <p>
                Thanks for using this extension! This extension will help you transfer urgent text to your PC without the hassle of using emails or any other app. Simply install the "Send it to your phone!" mobile app and create a connection between both devices to get started!
            </p>
            <button onClick={handleClick}>Connect</button>
        </div>
    );
};

export default Thanks;