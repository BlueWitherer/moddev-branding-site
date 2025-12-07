import './index.css';

import { BrowserRouter, Route, Routes } from 'react-router-dom';
import { createRoot } from 'react-dom/client'
import { StrictMode } from 'react';

import Login from './Login.tsx';
import NotFound from './NotFound.tsx';
import Dashboard from './Dashboard.tsx';

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<Login />} />
                <Route path="/dashboard" element={<Dashboard />} />

                <Route path="*" element={<NotFound />} />
            </Routes>
        </BrowserRouter>
    </StrictMode>
);