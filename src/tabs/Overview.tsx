import { useState, useEffect } from 'react';

import type { Image, User } from "../Include.mts";

import { Box, Paper, Typography, Grid, Card, CardMedia, CardContent, Chip } from "@mui/material";

import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import HourglassEmptyIcon from '@mui/icons-material/HourglassEmpty';

interface OverviewProps {
    user: User | null;
};

function Overview({ user }: OverviewProps) {
    const [images, setImages] = useState<Image[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchImages = async () => {
            try {
                const res = await fetch('/brand/list');
                if (res.ok) {
                    const data = await res.json();
                    setImages(data || []);
                } else {
                    console.error("Failed to fetch user images");
                };
            } catch (error) {
                console.error(error);
            } finally {
                setLoading(false);
            };
        };

        if (user) fetchImages();
    }, [user]);

    return (
        <Box sx={{ maxWidth: 1000, mx: 'auto', p: 3 }}>
            <Typography variant="h4" gutterBottom sx={{ mb: 4, textAlign: 'center', fontFamily: "'Russo One', sans-serif" }}>
                Hello, {user?.login}!
            </Typography>

            <Paper sx={{ p: 4, bgcolor: 'rgba(0,0,0,0.4)', color: 'white', minHeight: 200 }}>
                <Typography variant="h6" gutterBottom sx={{ mb: 3, textAlign: 'left', borderBottom: '1px solid rgba(255,255,255,0.1)', pb: 1 }}>
                    Your Branding
                </Typography>
                {loading ? (
                    <Typography textAlign="center">Loading...</Typography>
                ) : images.length === 0 ? (
                    <Typography textAlign="center" sx={{ color: 'rgba(255,255,255,0.7)', py: 4 }}>
                        You haven't submitted any branding images yet.
                        Go to the Submission tab to upload one!
                    </Typography>
                ) : (
                    <Grid container spacing={3} sx={{ justifyContent: 'center', alignItems: 'center' }}>
                        {images.map((img) => (
                            <Grid size={{ xs: 12, sm: 6, md: 4 }} key={img.id}>
                                <Card sx={{
                                    bgcolor: 'rgba(255, 255, 255, 0.05)',
                                    border: '1px solid rgba(255, 255, 255, 0.1)',
                                    height: '100%',
                                    display: 'flex',
                                    flexDirection: 'column',
                                }}>
                                    <Box sx={{ position: 'relative', pt: '56.25%' /* 16:9 aspect ratio */ }}>
                                        <CardMedia
                                            component="img"
                                            image={img.image_url}
                                            alt={`Branding ${img.id}`}
                                            sx={{
                                                position: 'absolute',
                                                top: 0,
                                                left: 0,
                                                width: '100%',
                                                height: '100%',
                                                objectFit: 'contain',
                                                p: 1,
                                                bgcolor: 'rgba(0,0,0,0.2)'
                                            }}
                                        />
                                        <Box sx={{ position: 'absolute', top: 8, right: 8 }}>
                                            {img.pending ? (
                                                <Chip
                                                    icon={<HourglassEmptyIcon sx={{ color: 'white !important' }} />}
                                                    label="Pending"
                                                    color="warning"
                                                    size="small"
                                                    sx={{ color: 'white' }}
                                                />
                                            ) : (
                                                <Chip
                                                    icon={<CheckCircleIcon sx={{ color: 'white !important' }} />}
                                                    label="Approved"
                                                    color="success"
                                                    size="small"
                                                    sx={{ color: 'white' }}
                                                />
                                            )}
                                        </Box>
                                    </Box>
                                    <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', gap: 1 }}>
                                        <Typography variant="caption" sx={{ color: 'rgba(255,255,255,0.6)' }}>
                                            Submitted: {new Date(img.created_at || '').toLocaleDateString()}
                                        </Typography>
                                    </CardContent>
                                </Card>
                            </Grid>
                        ))}
                    </Grid>
                )}
            </Paper>
        </Box>
    );
};

export default Overview;