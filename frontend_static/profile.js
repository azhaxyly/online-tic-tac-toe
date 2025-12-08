// --- Profile Functions ---

async function fetchUserProfile(nickname) {
    try {
        const res = await fetch(`${API_URL}/api/profile/${encodeURIComponent(nickname)}`, { credentials: 'include' });
        if (!res.ok) throw new Error('Failed to fetch profile');
        return await res.json();
    } catch (err) {
        console.error('Profile fetch error:', err);
        return null;
    }
}

async function showProfileModal() {
    const modal = document.getElementById('profile-modal');
    modal.classList.remove('hidden');

    try {
        const res = await fetch(`${API_URL}/api/profile-stats`, { credentials: 'include' });
        if (!res.ok) throw new Error('Failed to fetch profile');
        const data = await res.json();

        document.getElementById('profile-nickname').textContent = data.nickname;
        document.getElementById('profile-elo').textContent = data.elo_rating || 1000;
        document.getElementById('profile-wins').textContent = data.wins || 0;
        document.getElementById('profile-losses').textContent = data.losses || 0;
        document.getElementById('profile-draws').textContent = data.draws || 0;

        const totalGames = (data.wins || 0) + (data.losses || 0) + (data.draws || 0);
        const winrate = totalGames > 0 ? Math.round((data.wins / totalGames) * 100) : 0;
        document.getElementById('profile-winrate').textContent = `Winrate: ${winrate}%`;
    } catch (err) {
        console.error('Error loading profile:', err);
        document.getElementById('profile-nickname').textContent = 'Error loading profile';
    }
}

function hideProfileModal() {
    document.getElementById('profile-modal').classList.add('hidden');
}

async function showOpponentProfileModal() {
    if (!opponentNickname) return;

    const modal = document.getElementById('opponent-profile-modal');
    modal.classList.remove('hidden');

    const data = await fetchUserProfile(opponentNickname);
    if (data) {
        document.getElementById('opp-profile-nickname').textContent = data.nickname;
        document.getElementById('opp-profile-elo').textContent = data.elo_rating || 1000;
        document.getElementById('opp-profile-wins').textContent = data.wins || 0;
        document.getElementById('opp-profile-losses').textContent = data.losses || 0;
        document.getElementById('opp-profile-draws').textContent = data.draws || 0;

        const totalGames = (data.wins || 0) + (data.losses || 0) + (data.draws || 0);
        const winrate = totalGames > 0 ? Math.round((data.wins / totalGames) * 100) : 0;
        document.getElementById('opp-profile-winrate').textContent = `Winrate: ${winrate}%`;
    } else {
        document.getElementById('opp-profile-nickname').textContent = 'Error loading profile';
    }
}

function hideOpponentProfileModal() {
    document.getElementById('opponent-profile-modal').classList.add('hidden');
}

function showPlayerPanels() {
    document.getElementById('opponent-panel').classList.remove('hidden');
    document.getElementById('my-panel').classList.remove('hidden');
}

function hidePlayerPanels() {
    document.getElementById('opponent-panel').classList.add('hidden');
    document.getElementById('my-panel').classList.add('hidden');
}

async function updatePlayerPanels() {
    // Update my panel
    const myData = await fetchUserProfile(myNickname);
    if (myData) {
        document.getElementById('my-name').textContent = myNickname;
        document.getElementById('my-elo').textContent = `ELO: ${myData.elo_rating || 1000}`;
    }

    // Update opponent panel
    if (opponentNickname && !opponentNickname.startsWith('Bot_')) {
        const oppData = await fetchUserProfile(opponentNickname);
        if (oppData) {
            document.getElementById('opponent-name').textContent = opponentNickname;
            document.getElementById('opponent-elo').textContent = `ELO: ${oppData.elo_rating || 1000}`;
        }
    } else if (opponentNickname) {
        document.getElementById('opponent-name').textContent = opponentNickname;
        document.getElementById('opponent-elo').textContent = 'Bot';
    }
}

// Initialize profile event handlers
document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('profile-btn').addEventListener('click', showProfileModal);
    document.getElementById('close-profile').addEventListener('click', hideProfileModal);
    document.getElementById('profile-modal').addEventListener('click', (e) => {
        if (e.target === document.getElementById('profile-modal')) {
            hideProfileModal();
        }
    });

    document.getElementById('close-opponent-profile').addEventListener('click', hideOpponentProfileModal);
    document.getElementById('opponent-profile-modal').addEventListener('click', (e) => {
        if (e.target === document.getElementById('opponent-profile-modal')) {
            hideOpponentProfileModal();
        }
    });

    document.getElementById('opponent-panel').addEventListener('click', showOpponentProfileModal);
});
