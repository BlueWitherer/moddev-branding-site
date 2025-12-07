export interface User {
    id: number; // GitHub user ID
    login: string; // GitHub username
    avatar_url?: string; // GitHub user avatar URL
    is_admin?: boolean; // Active administrator status
    is_staff?: boolean; // Active staff status
    verified?: boolean; // Trusted status
    banned?: boolean; // Banned status
    created_at?: string; // First created
    updated_at?: string; // Last updated
};

export interface Image {
    img_id: string;
    user_id: string;
    image_url: string;
    created_at?: string;
    pending?: boolean;
};