import { useState, useEffect } from 'react';
import { Box, Paper, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Button, Snackbar, Alert } from "@mui/material";
import CheckCircleIcon from '@mui/icons-material/CheckCircle';

interface Img {
    id: number;
    user_id: number;
    image_url: string;
    created_at: string;
    pending: boolean;
    login: string;
}

function Pending() {
    const [images, setImages] = useState<Img[]>([]);
    const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

    const fetchImages = async () => {
        try {
            const res = await fetch('/brand/pending');
            if (res.ok) {
                const data = await res.json();
                setImages(data);
            } else {
                console.error("Failed to fetch pending images");
            }
        } catch (error) {
            console.error(error);
        }
    };

    useEffect(() => {
        fetchImages();
    }, []);

    const handleAccept = async (id: number) => {
        try {
            const res = await fetch(`/brand/pending/accept?id=${id}`, {
                method: 'POST'
            });
            if (res.ok) {
                setMessage({ type: 'success', text: 'Image accepted successfully!' });
                fetchImages();
            } else {
                const errorText = await res.text();
                setMessage({ type: 'error', text: `Failed to accept: ${errorText}` });
            }
        } catch (error) {
            setMessage({ type: 'error', text: 'An unexpected error occurred.' });
            console.error(error);
        }
    };

    const handleCloseMessage = () => {
        setMessage(null);
    };

    return (
        <Box sx={{ maxWidth: 1000, mx: 'auto', p: 3 }}>
            <Typography variant="h4" gutterBottom sx={{ mb: 4, textAlign: 'center', fontFamily: "'Russo One', sans-serif" }}>
                Pending Brandings
            </Typography>
            <TableContainer component={Paper} sx={{ bgcolor: 'rgba(0,0,0,0.4)', color: 'white' }}>
                <Table>
                    <TableHead>
                        <TableRow>
                            <TableCell sx={{ color: 'white', fontWeight: 'bold' }}>ID</TableCell>
                            <TableCell sx={{ color: 'white', fontWeight: 'bold' }}>User ID</TableCell>
                            <TableCell sx={{ color: 'white', fontWeight: 'bold' }}>Username</TableCell>
                            <TableCell sx={{ color: 'white', fontWeight: 'bold' }}>Image</TableCell>
                            <TableCell sx={{ color: 'white', fontWeight: 'bold' }}>Created At</TableCell>
                            <TableCell sx={{ color: 'white', fontWeight: 'bold' }}>Action</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {images.map((img) => (
                            <TableRow key={img.id} hover sx={{ '&:hover': { bgcolor: 'rgba(255,255,255,0.05)' } }}>
                                <TableCell sx={{ color: 'white' }}>{img.id}</TableCell>
                                <TableCell sx={{ color: 'white' }}>{img.user_id}</TableCell>
                                <TableCell sx={{ color: 'white' }}>{img.login}</TableCell>
                                <TableCell sx={{ color: 'white' }}>
                                    <Box
                                        component="img"
                                        src={img.image_url}
                                        alt="branding"
                                        sx={{
                                            maxWidth: 100,
                                            maxHeight: 60,
                                            borderRadius: 1,
                                            cursor: 'pointer',
                                            border: '1px solid rgba(255,255,255,0.1)'
                                        }}
                                        onClick={() => window.open(img.image_url, '_blank')}
                                    />
                                </TableCell>
                                <TableCell sx={{ color: 'white' }}>
                                    {new Date(img.created_at).toLocaleString()}
                                </TableCell>
                                <TableCell sx={{ color: 'white' }}>
                                    <Button
                                        variant="contained"
                                        color="success"
                                        size="small"
                                        startIcon={<CheckCircleIcon />}
                                        onClick={() => handleAccept(img.id)}
                                        sx={{ textTransform: 'none' }}
                                    >
                                        Accept
                                    </Button>
                                </TableCell>
                            </TableRow>
                        ))}
                        {images.length === 0 && (
                            <TableRow>
                                <TableCell colSpan={5} align="center" sx={{ color: 'rgba(255,255,255,0.7)', py: 4 }}>
                                    No pending images found.
                                </TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </TableContainer>

            <Snackbar open={!!message} autoHideDuration={6000} onClose={handleCloseMessage}>
                {message ? (
                    <Alert onClose={handleCloseMessage} severity={message.type} sx={{ width: '100%' }}>
                        {message.text}
                    </Alert>
                ) : undefined}
            </Snackbar>
        </Box>
    );
};

export default Pending;