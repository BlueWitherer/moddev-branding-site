import "./Log.mts";

import './App.css';

import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

import GitHubIcon from '@mui/icons-material/GitHub';

function App() {
    const navigate = useNavigate();

    useEffect(() => {
        fetch("/session", { credentials: "include" })
            .then((res) => {
                if (res.ok) {
                    navigate("/dashboard");
                }
            })
            .catch();
    }, [navigate]);

    return (
        <div className="container" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', minHeight: '100vh' }}>
            <div>
                <a href='https://www.geode-sdk.org/mods/cheeseworks.moddevbranding' target="_blank">
                    <img src='/icon.png' className="logo" alt="Mod Developer Branding logo" />
                </a>
            </div>
            <h1>Mod Developer Branding</h1>
            <p>Add your branding to Geode's mod information popups in-game!</p>
            <div>
                <button onClick={() => navigate('/dashboard')}>
                    <GitHubIcon /> Login
                </button>
            </div>
        </div>
    );
};

export default App;