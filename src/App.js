
import React, { memo, useEffect, useState } from 'react'
import CreateConnection from './components/CreateConnection'
import ConnectionConnected from './components/ConnectionConnected'
import Thanks from './components/Thanks'

import axios from 'axios';
import HaveACode from './components/HaveACode';

// "undefined" means the URL will be computed from the `window.location` object
const URL = process.env.NODE_ENV === 'production' ? undefined : 'http://localhost:3000';

var socket = null;

const App = memo(() => {
    const [connectionState, setConnectionState] = useState({
        status: 'thanks',
        to: null,
        secretCode: "123456"
    })

    const handleConnect = () => {
        // Code to create WebSocket connection
        console.log('Connecting to secret code: ', connectionState.secretCode);

        setConnectionState({
            ...connectionState,
            status: 'connecting'
        })

        axios.post('/confirm', { secretCode: connectionState.secretCode })
            .then(response => {
                // handle success
                console.log(response.data);
            })
            .catch(error => {
                // handle error
                console.error(error);
            });

        // if ( socket && socket.connected ) {
        //     socket.off('connect', onConnect);
        //     socket.off('disconnect', onDisconnect);

        //     socket.disconnect();
        // }

        // socket = io(URL, { query: { type: 'sender', code: secretCode } });
        // socket.on('connect', onConnect);
        // socket.on('disconnect', onDisconnect);
    };

    const onConnect = () => {
        setConnectionState({
            status: 'connected',
            to: connectionState.secretCode
        })
    }

    const onDisconnect = () => {
        setConnectionState({
            status: 'thanks',
            to: null
        })
    }

    const handleDisconnect = () => {
        if ( socket && socket.connected ) {
            socket.off('connect', onConnect);
            socket.off('disconnect', onDisconnect);
            socket.disconnect();
        }
        // Logic to disconnect the websocket
        setConnectionState({ ...connectionState, status: 'thanks' });
    };

    const handleSendMessage = (msg) => {
        // Logic to send the message
        alert('Message sent!');
        
        socket.emit("message", { code: connectionState.secretCode, message: msg})
    }

    useEffect(() => {
        return () => {
            if ( socket && socket.connected ) {
                socket.off('connect', onConnect);
                socket.off('disconnect', onDisconnect);
                socket.disconnect();
            }
        }
    }, [])

    const setSecretCode = (secretCode) => setConnectionState({...connectionState, to: secretCode})
    
    return (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
            {connectionState.status === 'disconnected' && <CreateConnection handleConnect={handleConnect} secretCode={connectionState.secretCode} setSecretCode={setSecretCode} connectionState={connectionState} setConnectionState={setConnectionState} />}
            {connectionState.status === 'connecting' && <p>Connecting...</p>}
            {connectionState.status === 'connected' && <ConnectionConnected handleSendMessage={handleSendMessage} handleDisconnect={handleDisconnect} connectionState={connectionState} setConnectionState={setConnectionState} />}
            {connectionState.status === 'thanks' && <Thanks connectionState={connectionState} setConnectionState={setConnectionState} />}
            {connectionState.status === 'haveACode' && <HaveACode connectionState={connectionState} setConnectionState={setConnectionState} />}
        </div>
    )
})

export default App
