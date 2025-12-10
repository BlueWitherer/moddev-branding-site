import { Box, Paper, Typography } from "@mui/material";

function Pending() {
    return (
        <Box sx={{ maxWidth: 800, mx: 'auto', p: 3 }}>
            <Typography variant="h4" gutterBottom sx={{ mb: 4, textAlign: 'center', fontFamily: "'Russo One', sans-serif" }}>
                Pending
            </Typography>
            <Paper sx={{ p: 4, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3, bgcolor: 'rgba(0,0,0,0.4)', color: 'white' }}>
                <Typography variant="body1" textAlign="center">
                    {/* TODO: pending thing pls */}
                    soon
                </Typography>
            </Paper>
        </Box>
    );
};

export default Pending;