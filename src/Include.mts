export interface User {
    /** GitHub user ID */
    id: number;
    /** GitHub username */
    login: string;
    /** GitHub user avatar URL */
    avatar_url?: string;
    /** Active administrator status */
    is_admin?: boolean;
    /** Active staff status */
    is_staff?: boolean;
    /** Trusted status */
    verified?: boolean;
    /** Banned status */
    banned?: boolean;
    /** First created */
    created_at?: string;
    /** Last updated */
    updated_at?: string;
};

export interface Image {
    /** Brand image ID */
    id: string;
    /** GitHub user ID */
    user_id: string;
    /** Brand image URL */
    image_url: string;
    /** First created */
    created_at?: string;
    /** Under review */
    pending?: boolean;
};