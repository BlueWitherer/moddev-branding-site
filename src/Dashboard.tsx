import "./Log.mts";

import { useEffect, useState } from "react";
import type { SyntheticEvent } from "react";
import { useNavigate } from "react-router-dom";

import type { User } from "./Include.mts";
import { Avatar, Box, Tabs, Tab, Typography, IconButton, useTheme, useMediaQuery, Select, MenuItem, FormControl, Button, Dialog, DialogTitle, DialogContent } from "@mui/material";

import type { SelectChangeEvent } from "@mui/material";

import LogoutIcon from '@mui/icons-material/Logout';
import Diversity1Icon from '@mui/icons-material/Diversity1';

import YouTubeIcon from '@mui/icons-material/YouTube';
import GitHubIcon from '@mui/icons-material/GitHub';
import XIcon from '@mui/icons-material/X';

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
                <Box sx={{ p: { xs: 2, md: 3 } }}>
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
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

    const [user, setUser] = useState<User | null>(null);
    const [tabValue, setTabValue] = useState(0);

    const handleTabChange = (_event: SyntheticEvent, newValue: number) => {
        setTabValue(newValue);
    };

    const handleSelectChange = (event: SelectChangeEvent<number>) => {
        setTabValue(Number(event.target.value));
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

    const [creditsOpen, setCreditsOpen] = useState(false);

    const handleOpenCredits = () => setCreditsOpen(true);
    const handleCloseCredits = () => setCreditsOpen(false);

    const showPending = user?.is_admin || user?.is_staff;

    return (
        <>
            <Box sx={{ width: '100%', bgcolor: 'rgba(0, 0, 0, 0.5)', position: 'relative' }}>
                <Box sx={{ borderBottom: 1, borderColor: 'divider', display: 'flex', justifyContent: 'center', p: isMobile ? 2 : 0 }}>
                    {isMobile ? (
                        <FormControl fullWidth >
                            <Select
                                value={tabValue}
                                onChange={handleSelectChange}
                                label="Navigation"
                                className="dashboard-select"
                                MenuProps={{
                                    PaperProps: {
                                        className: 'dashboard-menu-paper'
                                    }
                                }}
                            >
                                <MenuItem value={0}>Dashboard</MenuItem>
                                <MenuItem value={1}>Submission</MenuItem>
                                {showPending && <MenuItem value={2}>Pending</MenuItem>}
                                <MenuItem value={3}>Settings</MenuItem>
                            </Select>
                        </FormControl>
                    ) : (
                        <Tabs
                            value={tabValue}
                            onChange={handleTabChange}
                            aria-label="dashboard tabs"
                            centered
                            className="custom-tabs"
                        >
                            <Tab label="Dashboard" value={0} {...a11yProps(0)} />
                            <Tab label="Submission" value={1} {...a11yProps(1)} />
                            {showPending && <Tab label="Pending" value={2} {...a11yProps(2)} />}
                            <Tab label="Settings" value={3} {...a11yProps(3)} />
                        </Tabs>
                    )}
                </Box>
            </Box>

            <CustomTabPanel value={tabValue} index={0}>
                <Overview user={user} />
            </CustomTabPanel>
            <CustomTabPanel value={tabValue} index={1}>
                <Submission />
            </CustomTabPanel>
            {showPending && (
                <CustomTabPanel value={tabValue} index={2}>
                    <Pending />
                </CustomTabPanel>
            )}
            <CustomTabPanel value={tabValue} index={3}>
                <Settings user={user} />
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
                zIndex: 1000,
                width: isMobile ? 'calc(100% - 120px)' : 'auto', // Reduced width on mobile to make space for credits button
                maxWidth: '100%',
                boxSizing: 'border-box'
            }}>
                <Avatar src={user?.avatar_url} />
                <Typography variant="body1" sx={{
                    color: 'white',
                    fontWeight: 'bold',
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    flexGrow: 1,
                    minWidth: 0
                }}>
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

            {/* Credits Button */}
            <Box
                sx={{
                    position: 'fixed',
                    bottom: 20,
                    right: 20,
                    zIndex: 1000,
                }}
            >
                <Button
                    onClick={handleOpenCredits}
                    sx={{
                        bgcolor: 'rgba(0, 0, 0, 0.8)',
                        backdropFilter: 'blur(10px)',
                        color: 'white',
                        '&:hover': {
                            bgcolor: 'rgba(0, 0, 0, 0.9)',
                        },
                        borderRadius: '24px', // Squircle shape
                        width: '72px',
                        height: '72px',
                        minWidth: 0,
                        p: 0,
                        fontWeight: 'bold',
                        fontSize: '0.75rem',
                    }}
                >
                    <Diversity1Icon />
                </Button>
            </Box>

            {/* Credits Modal */}
            <Dialog
                open={creditsOpen}
                onClose={handleCloseCredits}
                aria-labelledby="credits-modal-title"
                aria-describedby="credits-modal-description"
                slotProps={{
                    paper: {
                        sx: {
                            bgcolor: 'rgba(20, 20, 20, 0.95)', // Dark background
                            border: '2px solid rgba(253, 128, 236, 1)',
                            boxShadow: 24,
                            borderRadius: 2,
                            color: 'white',
                            width: isMobile ? '80%' : 500,
                            maxWidth: 'none', // Allow custom width
                        }
                    }
                }}
            >
                <DialogTitle id="credits-modal-title" sx={{ mb: 1, color: 'rgba(255, 255, 255, 1)', fontFamily: "'Russo One', sans-serif", textAlign: 'center' }}>
                    Credits
                </DialogTitle>

                <DialogContent>
                    <Box sx={{ display: 'flex', justifyContent: 'space-around', alignItems: 'flex-start', mt: 1 }}>
                        {/* ArcticWoof */}
                        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                            <Avatar
                                src="https://avatars.githubusercontent.com/u/56347227"
                                sx={{ width: '70%', height: '70%', mb: 1, cursor: 'pointer' }}
                                onClick={() => window.open("https://github.com/DumbCaveSpider", "_blank")}
                            />
                            <Typography variant="h6" sx={{ fontWeight: 'bold' }}>ArcticWoof</Typography>
                            <Typography variant="body2" sx={{ color: 'rgba(255, 255, 255, 0.7)' }}>
                                Frontend/UI/UX
                            </Typography>
                        </Box>

                        {/* Cheeseworks */}
                        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                            <Avatar
                                src="https://avatars.githubusercontent.com/u/47698640"
                                sx={{ width: '70%', height: '70%', mb: 1, cursor: 'pointer' }}
                                onClick={() => window.open("https://github.com/BlueWitherer", "_blank")}
                            />
                            <Typography variant="h6" sx={{ fontWeight: 'bold' }}>Cheeseworks</Typography>
                            <Typography variant="body2" sx={{ color: 'rgba(255, 255, 255, 0.7)' }}>
                                Backend/API
                            </Typography>
                            <Typography variant="body2" sx={{ color: 'rgba(255, 255, 255, 0.7)' }}>
                                Geode Mod
                            </Typography>
                        </Box>
                    </Box>
                    <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', mt: 1 }}>
                        <IconButton
                            className="social-button"
                            component="a"
                            href="https://www.youtube.com/@cheese_works/"
                            target="_blank"
                            rel="noopener noreferrer"
                            aria-label="YouTube"
                        >
                            <YouTubeIcon />
                        </IconButton>

                        <IconButton
                            className="social-button"
                            component="a"
                            href="https://www.github.com/BlueWitherer/"
                            target="_blank"
                            rel="noopener noreferrer"
                            aria-label="GitHub"
                        >
                            <GitHubIcon />
                        </IconButton>

                        <IconButton
                            className="social-button"
                            component="a"
                            href="https://www.x.com/chris_rhatt/"
                            target="_blank"
                            rel="noopener noreferrer"
                            aria-label="X"
                        >
                            <XIcon />
                        </IconButton>
                    </Box>
                    <Button variant="contained" color="error" onClick={handleCloseCredits} sx={{ mt: 2, width: '100%' }}>
                        Close
                    </Button>
                </DialogContent>
            </Dialog >
        </>
    );
};

export default Dashboard;