import "./Log.mts";

import './index.css';

import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

import GitHubIcon from '@mui/icons-material/GitHub';

function App() {
    const navigate = useNavigate();

    const handleLogin = () => {
        fetch("/session", { credentials: "include" })
            .then((res) => {
                if (res.ok) {
                    navigate("/dashboard");
                } else {
                    console.warn("User not logged in");
                    window.location.href = "/login";
                };
            })
            .catch(() => {
                console.warn("Session invalid");
                window.location.href = "/login";
            });
    };

    useEffect(() => {
        fetch("/session", { credentials: "include" })
            .then((res) => {
                if (res.ok) {
                    navigate("/dashboard");
                } else {
                    console.error("User not logged in");
                };
            })
            .catch(() => {
                console.error("Session invalid");
            });
    }, [navigate]);

    return (
        <div className="container" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', textAlign: 'center' }}>
            <div>
                <a href='https://www.geode-sdk.org/mods/cheeseworks.moddevbranding' target="_blank">
                    <img src='/icon.png' className="logo" alt="Mod Developer Branding logo" />
                </a>
            </div>
            <h1>Mod Developer Branding</h1>
            <p>Add your branding to Geode's mod information popups in-game!</p>
            <div>
                <button onClick={handleLogin}>
                    <GitHubIcon /> Login
                </button>
            </div>
        </div>
    );
};

export default App;