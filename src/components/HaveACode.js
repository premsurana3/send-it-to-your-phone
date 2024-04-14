import axios from 'axios';
import React, { memo, useEffect, useState } from 'react'

const HaveACode = memo(({connectionState, setConnectionState}) => {

    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        axios.post("/request",{ secretCode: connectionState.secretCode })
            .then((response) => {
                console.log(response.data);
            })
            .catch((error) => {
                console.error(error);
                setError("An error occurred while trying to connect");
            }).finally(() => {
                setLoading(false);
            })
    }, [])

    if ( error ) {
        return <p>Error...</p>
    }

    if ( loading ) {
        return <p>loading....</p>
    }

    return (
        <div>HaveACode: {connectionState.secretCode}</div>
    )
})

export default HaveACode