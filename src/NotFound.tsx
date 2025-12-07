import "./Log.mts";

import { useNavigate } from "react-router-dom";

export default function NotFound() {
    console.error("404 - Page Not Found");
    const navigate = useNavigate();

    const onBack = () => {
        console.info("Navigating back to home page");
        navigate("/");
    };

    return (
        <div className="container" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', minHeight: '100vh', alignItems: 'center' }}>
            <h1 style={{ marginTop: '2rem', marginBottom: '1rem' }}>404</h1>
            <h6 style={{ marginBottom: '1rem' }}>Oops! Couldn't find that page...</h6>

            <button onClick={onBack}>
                Go Back
            </button>
        </div>
    );
}