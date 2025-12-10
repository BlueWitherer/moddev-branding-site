import type { User } from "../Include.mts";
import { Box, Paper, Typography } from "@mui/material";

interface OverviewProps {
    user: User | null;
}

function Overview({ user }: OverviewProps) {
    return (
        <Box sx={{ maxWidth: 800, mx: 'auto', p: 3 }}>
            <Typography variant="h4" gutterBottom sx={{ mb: 4, textAlign: 'center', fontFamily: "'Russo One', sans-serif" }}>
                Hello! {user?.login}
            </Typography>
            <Paper sx={{ p: 4, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3, bgcolor: 'rgba(0,0,0,0.4)', color: 'white' }}>
                <Typography variant="body1" textAlign="center">
                    shows the mod dev's branding image here
                </Typography>
            </Paper>
        </Box>
    );
};

export default Overview;