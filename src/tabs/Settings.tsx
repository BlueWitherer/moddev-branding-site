import { useState } from "react";

import type { User } from "../Include.mts";

import { Box, Button, Typography, Paper, Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions } from "@mui/material";

import { SiGeode } from "react-icons/si";
import DeleteForeverIcon from '@mui/icons-material/DeleteForever';

interface SettingsProps {
    user: User | null;
}

function Settings({ user }: SettingsProps) {
    const [open, setOpen] = useState(false);

    {/* TODO: use delete endpoint to delete it ig*/ }
    const handleDeleteOpen = () => {
        setOpen(true);
    };

    const handleDeleteClose = () => {
        setOpen(false);
    };

    return (
        <Box sx={{ maxWidth: 800, mx: 'auto', p: 3 }}>
            <Typography variant="h4" gutterBottom sx={{ mb: 4, textAlign: 'center', fontFamily: "'Russo One', sans-serif" }}>
                Settings
            </Typography>

            <Paper sx={{ p: 4, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3, bgcolor: 'rgba(0,0,0,0.4)', color: 'white' }}>
                <Box sx={{ textAlign: 'center', width: '100%' }}>
                    <Typography variant="h5" gutterBottom sx={{ mb: 2, textAlign: 'center', fontFamily: "'Russo One', sans-serif" }}>
                        Account Information
                    </Typography>
                    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1, my: 2 }}>
                        <Typography variant="body1">
                            <strong>User ID:</strong> {user?.id}
                        </Typography>
                        <Typography variant="body1">
                            <strong>Highest Role:</strong> {user?.is_admin ? "Admin" : user?.is_staff ? "Staff" : "User"}
                        </Typography>
                        <Typography variant="body1">
                            <strong>Verified:</strong> {user?.verified ? "Yes" : "No"}
                        </Typography>
                        <Typography variant="body1">
                            <strong>Created At:</strong> {user?.created_at?.toString()}
                        </Typography>
                    </Box>
                </Box>
            </Paper >
            <Paper sx={{ mt: 4, p: 4, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3, bgcolor: 'rgba(0,0,0,0.4)', color: 'white' }}>
                <Box sx={{ textAlign: 'center', width: '100%' }}>
                    <Typography variant="h5" gutterBottom sx={{ mb: 2, textAlign: 'center', fontFamily: "'Russo One', sans-serif", color: 'white' }}>
                        Geode Mod
                    </Typography>
                    <p>
                        Your mod developer branding can be seen in-game with the mod!
                    </p>
                    <DialogContent sx={{ p: 0, position: 'relative', display: 'inline-block' }}>
                        <img
                            src="https://www.github.com/BlueWitherer/ModDevBranding/blob/master/previews/preview-1.png?raw=true"
                            alt="Mod Preview"
                            style={{ display: 'block', maxWidth: '100%', maxHeight: '90vh' }}
                        />
                    </DialogContent>
                    <Box sx={{ textAlign: 'center', width: '100%' }}>
                        <Button
                            variant="contained"
                            sx={{
                                mt: 2,
                                px: 5,
                                bgcolor: 'rgb(253, 128, 241)',
                                '&:hover': {
                                    bgcolor: 'rgb(200, 100, 190)'
                                },
                                '&.Mui-disabled': {
                                    bgcolor: 'rgba(253, 128, 241, 0.3)',
                                    color: 'rgba(255, 255, 255, 0.5)'
                                }
                            }}
                            href="https://www.geode-sdk.org/mods/cheeseworks.moddevbranding"
                            target="_blank">
                            <SiGeode className="simple-icon" /> Download
                        </Button>
                    </Box>
                </Box>
            </Paper >
            <Paper sx={{ mt: 4, p: 4, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3, bgcolor: 'rgba(0,0,0,0.4)', color: 'white' }}>
                <Box sx={{ textAlign: 'center', width: '100%' }}>
                    <Typography variant="h5" gutterBottom sx={{ mb: 2, textAlign: 'center', fontFamily: "'Russo One', sans-serif", color: 'rgba(253, 128, 128, 1)' }}>
                        Dangerous Actions
                    </Typography>
                    <Box sx={{ textAlign: 'center', width: '100%' }}>
                        <Button variant="contained" color="error" onClick={handleDeleteOpen}>
                            <DeleteForeverIcon /> Delete Account
                        </Button>
                    </Box>
                </Box>
            </Paper >

            <Dialog
                open={open}
                onClose={handleDeleteClose}
                aria-labelledby="alert-dialog-title"
                aria-describedby="alert-dialog-description"
                slotProps={{
                    paper: {
                        sx: {
                            bgcolor: 'rgba(20, 20, 20, 0.95)',
                            color: 'white',
                            backdropFilter: 'blur(10px)',
                            border: '1px solid rgba(253, 128, 128, 1)',
                            borderRadius: 2
                        }
                    }
                }}
            >
                <DialogTitle id="alert-dialog-title" sx={{ fontFamily: "'Russo One', sans-serif", color: 'rgba(253, 128, 128, 1)' }}>
                    {"Delete Account?"}
                </DialogTitle>
                <DialogContent>
                    <DialogContentText id="alert-dialog-description" sx={{ color: 'rgba(255, 255, 255, 0.7)' }}>
                        Are you sure you want to delete your account? This action cannot be undone.
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={handleDeleteClose} sx={{ color: 'white' }}>Cancel</Button>
                    <Button onClick={handleDeleteClose} color="error" autoFocus variant="contained" sx={{ bgcolor: 'rgba(253, 128, 128, 1)', color: 'black', '&:hover': { bgcolor: 'rgb(203, 78, 191)' } }}>
                        <DeleteForeverIcon /> Delete
                    </Button>
                </DialogActions>
            </Dialog>
        </Box >
    );
};

export default Settings;