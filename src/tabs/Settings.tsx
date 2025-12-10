import { Box, Typography, Paper } from "@mui/material";
import type { User } from "../Include.mts";

interface SettingsProps {
    user: User | null;
}

function Settings({ user }: SettingsProps) {
    return (
        <Box sx={{ maxWidth: 800, mx: 'auto', p: 3 }}>
            <Typography variant="h4" gutterBottom sx={{ mb: 4, textAlign: 'center', fontFamily: "'Russo One', sans-serif" }}>
                Settings
            </Typography>

            <Paper sx={{ p: 4, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3, bgcolor: 'rgba(0,0,0,0.4)', color: 'white' }}>
                <Box sx={{ textAlign: 'center', width: '100%' }}>
                    <Typography variant="h6" gutterBottom>
                        Account Information
                    </Typography>
                    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1, my: 2 }}>
                        <Typography variant="body1">
                            <strong>User ID:</strong> {user?.id}
                        </Typography>
                    </Box>
                </Box>
            </Paper>
        </Box>
    );
};

export default Settings;