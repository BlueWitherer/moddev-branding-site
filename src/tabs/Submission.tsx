import { useState, type ChangeEvent } from "react";
import { Box, Button, Typography, Paper, Alert, Snackbar, CircularProgress, Dialog, DialogContent } from '@mui/material';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';

function Submission() {
    const [file, setFile] = useState<File | null>(null);
    const [preview, setPreview] = useState<string | null>(null);
    const [uploading, setUploading] = useState(false);
    const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);
    const [openPreview, setOpenPreview] = useState(false);

    const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
        if (event.target.files && event.target.files[0]) {
            const selectedFile = event.target.files[0];
            setFile(selectedFile);
            setPreview(URL.createObjectURL(selectedFile));
        }
    };

    const handleSubmit = async () => {
        if (!file) return;

        setUploading(true);
        const formData = new FormData();
        formData.append('image-upload', file);

        try {
            const response = await fetch('/brand/submit', {
                method: 'POST',
                body: formData,
            });

            if (response.ok) {
                setMessage({ type: 'success', text: 'Brand image submitted successfully!' });
                setFile(null);
                setPreview(null);
            } else {
                const errorText = await response.text();
                setMessage({ type: 'error', text: `Upload failed: ${errorText}` });
            }
        } catch (error) {
            setMessage({ type: 'error', text: 'An unexpected error occurred.' });
            console.error(error);
        } finally {
            setUploading(false);
        }
    };

    const handleCloseMessage = () => {
        setMessage(null);
    };

    return (
        <Box sx={{ maxWidth: 800, mx: 'auto', p: 3 }}>
            <Typography variant="h4" gutterBottom sx={{ mb: 4, textAlign: 'center', fontFamily: "'Russo One', sans-serif" }}>
                Submit Branding
            </Typography>

            <Paper sx={{ p: 4, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3, bgcolor: 'rgba(0,0,0,0.4)', color: 'white' }}>
                <Typography variant="body1" textAlign="center">
                    Upload your custom branding image here. It will be reviewed by admins and staff.
                </Typography>

                <Button
                    component="label"
                    variant="outlined"
                    startIcon={<CloudUploadIcon />}
                    sx={{
                        p: 2,
                        borderStyle: 'dashed',
                        color: 'rgb(253, 128, 241)',
                        borderColor: 'rgb(253, 128, 241)',
                        '&:hover': {
                            borderColor: 'rgb(253, 128, 241)',
                            bgcolor: 'rgba(253, 128, 241, 0.1)'
                        }
                    }}
                >
                    Select Image
                    <input
                        type="file"
                        hidden
                        accept="image/*"
                        onChange={handleFileChange}
                    />
                </Button>

                {preview && (
                    <Box sx={{ mt: 2, textAlign: 'center', display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                        <Typography variant="subtitle2" gutterBottom>
                            Preview (Click to enlarge):
                        </Typography>
                        <Box
                            onClick={() => setOpenPreview(true)}
                            sx={{
                                position: 'relative',
                                display: 'inline-block',
                                maxHeight: { xs: 300, md: 500 },
                                maxWidth: '100%',
                                overflow: 'hidden',
                                cursor: 'pointer',
                                transition: 'transform 0.2s',
                                '&:hover': {
                                    transform: 'scale(1.02)'
                                }
                            }}
                        >
                            <img
                                src="/previewbg.png"
                                alt="Background"
                                style={{ display: 'block', maxWidth: '100%', maxHeight: 'inherit' }}
                            />
                            <img
                                src={preview}
                                alt="Preview"
                                style={{
                                    position: 'absolute',
                                    top: '18%',
                                    left: '38%',
                                    width: '60%',
                                    height: '78%',
                                    objectFit: 'contain',
                                    opacity: 0.2
                                }}
                            />
                        </Box>
                        <Typography variant="caption" display="block" sx={{ mt: 1 }}>
                            {file?.name}
                        </Typography>
                    </Box>
                )}

                <Button
                    variant="contained"
                    onClick={handleSubmit}
                    disabled={!file || uploading}
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
                >
                    {uploading ? <CircularProgress size={24} sx={{ color: 'white' }} /> : 'Submit Branding'}
                </Button>
            </Paper>

            <Dialog
                open={openPreview}
                onClose={() => setOpenPreview(false)}
                maxWidth="lg"
                fullWidth
                PaperProps={{
                    sx: {
                        bgcolor: 'transparent',
                        boxShadow: 'none',
                        overflow: 'hidden',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center'
                    }
                }}
            >
                <DialogContent sx={{ p: 0, position: 'relative', display: 'inline-block' }}>
                    <img
                        src="/previewbg.png"
                        alt="Full Background"
                        style={{ display: 'block', maxWidth: '100%', maxHeight: '90vh' }}
                    />
                    <img
                        src={preview || ''}
                        alt="Full Preview"
                        style={{
                            position: 'absolute',
                            top: '18%',
                            left: '38%',
                            width: '60%',
                            height: '78%',
                            objectFit: 'contain',
                            opacity: 0.2
                        }}
                    />
                </DialogContent>
            </Dialog>

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

export default Submission;