import "./Log.mts";

import './index.css';

import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

import { Box, Paper, Typography, Button } from '@mui/material';

import { SiGithub } from "react-icons/si";

function Login() {
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

    /**
     * Check if the user is already logged in
     */
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
        <Box
            sx={{
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'center',
                alignItems: 'center',
                minHeight: '100vh',
                p: 2
            }}
        >
            <Paper
                elevation={6}
                sx={{
                    p: 5,
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    textAlign: 'center',
                    bgcolor: 'rgba(0, 0, 0, 0.6)',
                    backdropFilter: 'blur(10px)',
                    borderRadius: 4,
                    maxWidth: 500,
                    width: '100%',
                    color: 'white'
                }}
            >
                <Box sx={{ mb: 3 }}>
                    <a href='https://www.geode-sdk.org/mods/cheeseworks.moddevbranding' target="_blank">
                        <img src='/icon.png' className="logo" alt="Mod Developer Branding logo" style={{ height: '8em' }} />
                    </a>
                </Box>

                <Typography variant="h4" component="h1" gutterBottom sx={{ fontWeight: 'bold' }}>
                    Mod Developer Branding
                </Typography>

                <Typography variant="body1" sx={{ mb: 4, color: 'rgba(255, 255, 255, 0.7)' }}>
                    Add your branding to Geode's mod information popups in-game!
                </Typography>

                <Button
                    variant="contained"
                    size="large"
                    onClick={handleLogin}
                    startIcon={<SiGithub />}
                    sx={{
                        borderRadius: '50px',
                        px: 5,
                        py: 1.5,
                        fontSize: '1.2rem',
                        textTransform: 'none',
                        backgroundColor: '#24292e',
                        '&:hover': {
                            backgroundColor: '#444c56'
                        }
                    }}
                >
                    Login with GitHub
                </Button>
            </Paper>
        </Box>
    );
};

export default Login;