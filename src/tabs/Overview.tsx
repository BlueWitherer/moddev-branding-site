import type { User } from "../Include.mts";

interface OverviewProps {
    user: User | null;
}

function Overview({ user }: OverviewProps) {
    return (
        <div className="container">
            <h1>Hello, {user?.login}!</h1>
            <p>Here's where you'll soon be able to manage your Geode mod developer branding!</p>
        </div>
    );
};

export default Overview;