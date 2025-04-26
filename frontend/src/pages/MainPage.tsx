import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

export default function MainPage() {
  const navigate = useNavigate();
  const [nickname, setNickname] = useState('');
  const [stats, setStats] = useState<{ online: number; active_games: number }>({ online: 0, active_games: 0 });
  const [error, setError] = useState('');

  useEffect(() => {
    // Получить nickname
    fetch('/api/nickname', { credentials: 'include' })
      .then(res => res.json())
      .then(data => setNickname(data.nickname))
      .catch(() => setError('Failed to fetch nickname.'));

    // Получить статистику сразу
    fetchStats();

    // Автообновление статистики раз в минуту
    const interval = setInterval(fetchStats, 60000);

    return () => clearInterval(interval);
  }, []);

  function fetchStats() {
    fetch('/api/stats')
      .then(res => res.json())
      .then(data => setStats(data))
      .catch(() => setError('Failed to fetch stats.'));
  }

  function handleQuickGame() {
    navigate('/online');
  }

  function handleOfflineGame() {
    navigate('/offline');
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-screen space-y-6">
      <h1 className="text-4xl font-bold">TicTacToe Online</h1>

      {error && <div className="text-red-500">{error}</div>}

      <div className="text-xl">Hello, <b>{nickname}</b>!</div>

      <div className="flex flex-col items-center space-y-2">
        {stats.online > 0 && <div>Players Online: {stats.online}</div>}
        {stats.active_games > 0 && <div>Active Games: {stats.active_games}</div>}
      </div>

      <div className="flex space-x-4 mt-4">
        <button
          onClick={handleQuickGame}
          className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          Quick Game
        </button>

        <button
          onClick={handleOfflineGame}
          className="px-6 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700"
        >
          Offline Game
        </button>
      </div>
    </div>
  );
}
