import "./Log.mts";

import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

import type { User } from "./Include.mts";
import { Avatar } from "@mui/material";

import AutoFixHighIcon from '@mui/icons-material/AutoFixHigh';

function Dashboard() {
    const navigate = useNavigate();

    const [user, setUser] = useState<User | null>(null);

    useEffect(() => {
        fetch("/session", { credentials: "include" })
            .then((res) => (res.ok ? res.json() : null))
            .then((u: User) => {
                console.debug("Received user information");

                if (u) {
                    console.debug("Processing user information...");
                    setUser({ ...u });
                    console.info(`Logged in as GitHub user ${u.login}!`);
                } else {
                    console.error("Invalid user");
                    navigate("/login");
                };
            })
            .catch((err: unknown) => {
                console.error(err);
                navigate("/");
            });
    }, [navigate]);

    return (
        <>
            <div className="centered">
                <Avatar src={user?.avatar_url} />
            </div>
            <div>
                <h4>Hello, {user?.login}!</h4>
                <p>Here's where you'll soon be able to manage your Geode mod developer branding!</p>
                <AutoFixHighIcon />
            </div>
        </>
    );
};

export default Dashboard;