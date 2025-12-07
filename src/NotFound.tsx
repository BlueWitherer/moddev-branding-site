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
        <>
            <div>
                <h1 className="text-6xl mt-8 mb-4">404</h1>
                <h6 className="text-1xl mb-4">Oops! Couldn't find that page...</h6>

                <button className="nine-slice-button" onClick={onBack}>
                    Go Back
                </button>
            </div>
        </>
    );
}