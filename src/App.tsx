import siteLogo from '../public/icon.png';
import './App.css'

function App() {
    return (
        <>
            <div>
                <a target="_blank">
                    <img src={siteLogo} className="logo" alt="Mod Developer Branding logo" />
                </a>
            </div>
            <h1>Mod Developer Branding</h1>
            <p className="read-the-docs">
                Coming soon...
            </p>
        </>
    );
};

export default App;