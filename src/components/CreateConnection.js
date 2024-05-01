import axios from 'axios';
import React, { useState, useMemo, useEffect } from 'react';

const CreateConnection = ({ connectionState, setConnectionState, secretCode, setSecretCode, handleConnect }) => {

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

const ShowAlreadyConnected = () => {
    const [connectedSessions, setConnectedSessions] = useState([]);

    useEffect(() => {

        const fetchConnectedSessions = async () => {
            try {
                const response = await axios.get('/api/findConnectedSessions');
                setConnectedSessions(response.data);
            } catch (error) {
                console.error('Error fetching connected sessions:', error);
            }
        }

        fetchConnectedSessions();
    }, []);
}

export default CreateConnection;