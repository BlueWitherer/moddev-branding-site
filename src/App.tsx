import "./Log.mts";

import './App.css';

import { useNavigate } from 'react-router-dom';

import GitHubIcon from '@mui/icons-material/GitHub';

function App() {
    const navigate = useNavigate();

    return (
        <>
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
        </>
    );
};

export default App;