import "./Log.mts";

import { useEffect, useState } from "react";
import type { SyntheticEvent } from "react";
import { useNavigate } from "react-router-dom";

import type { User } from "./Include.mts";
import { Avatar, Box, Tabs, Tab, Typography, IconButton } from "@mui/material";

import LogoutIcon from '@mui/icons-material/Logout';

import Overview from "./tabs/Overview";
import Submission from "./tabs/Submission";
import Pending from "./tabs/Pending";
import Settings from "./tabs/Settings";

interface TabPanelProps {
    children?: React.ReactNode;
    index: number;
    value: number;
};

function CustomTabPanel(props: TabPanelProps) {
    const { children, value, index, ...other } = props;

    return (
        <div
            role="tabpanel"
            hidden={value !== index}
            id={`simple-tabpanel-${index}`}
            aria-labelledby={`simple-tab-${index}`}
            {...other}
        >
            {value === index && (
                <Box sx={{ p: 3 }}>
                    {children}
                </Box>
            )}
        </div>
    );
};

function a11yProps(index: number) {
    return {
        "id": `simple-tab-${index}`,
        "aria-controls": `simple-tabpanel-${index}`,
    };
};

function Dashboard() {
    const navigate = useNavigate();

    const [user, setUser] = useState<User | null>(null);
    const [tabValue, setTabValue] = useState(0);

    const handleTabChange = (_event: SyntheticEvent, newValue: number) => {
        setTabValue(newValue);
    };

    const handleLogout = () => {
        fetch("/logout", { method: "POST" })
            .then(() => {
                console.warn("Logged out");
                window.location.href = "/";
            })
            .catch(console.error);
    };

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
                    window.location.href = "/login";
                };
            })
            .catch((err: unknown) => {
                console.error(err);
                window.location.href = "/login";
            });
    }, [navigate]);

    return (
        <>
            <Box sx={{ width: '100%', bgcolor: 'rgba(0, 0, 0, 0.5)', position: 'relative' }}>
                <Box sx={{ borderBottom: 1, borderColor: 'divider', display: 'flex', justifyContent: 'center' }}>
                    <Tabs
                        value={tabValue}
                        onChange={handleTabChange}
                        aria-label="dashboard tabs"
                        centered
                        className="custom-tabs"
                    >
                        <Tab label="Dashboard" {...a11yProps(0)} />
                        <Tab label="Submission" {...a11yProps(1)} />
                        <Tab label="Pending" {...a11yProps(2)} />
                        <Tab label="Settings" {...a11yProps(3)} />
                    </Tabs>
                </Box>
            </Box>

            <CustomTabPanel value={tabValue} index={0}>
                <Overview user={user} />
            </CustomTabPanel>
            <CustomTabPanel value={tabValue} index={1}>
                <Submission />
            </CustomTabPanel>
            <CustomTabPanel value={tabValue} index={2}>
                <Pending />
            </CustomTabPanel>
            <CustomTabPanel value={tabValue} index={3}>
                <Settings />
            </CustomTabPanel>

            <Box sx={{
                position: 'fixed',
                bottom: 20,
                left: 20,
                bgcolor: 'rgba(0, 0, 0, 0.8)',
                borderRadius: '16px',
                p: 2,
                display: 'flex',
                alignItems: 'center',
                gap: 2,
                zIndex: 1000
            }}>
                <Avatar src={user?.avatar_url} />
                <Typography variant="body1" sx={{ color: 'white', fontWeight: 'bold' }}>
                    {user?.login}
                </Typography>
                <IconButton
                    color="error"
                    onClick={handleLogout}
                    title="Logout"
                >
                    <LogoutIcon />
                </IconButton>
            </Box>
        </>
    );
};

export default Dashboard;