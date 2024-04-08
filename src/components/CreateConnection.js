import React, { useState, useMemo, useEffect } from 'react';

const CreateConnection = ({connectionState, setConnectionState, secretCode, setSecretCode, handleConnect}) => {

    const handleInputChange = (e) => {
        setSecretCode(e.target.value);
    }

    return (
        <div>
            <p>Enter Secret Code: <input type='text' value={secretCode} onChange={handleInputChange}></input></p>
            <button onClick={handleConnect}>Connect</button>
        </div>
    );
};

export default CreateConnection;