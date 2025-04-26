import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import MainPage from './pages/MainPage';
import OnlineGamePage from './pages/OnlineGamePage.tsx';
import OfflineGamePage from './pages/OfflineGamePage.tsx';

export default function App() {
  return (
    <Router>
      <div className="min-h-screen flex flex-col bg-gray-100">
        <nav className="p-4 bg-white shadow flex justify-between">
          <div className="font-bold text-xl">TicTacToe Online</div>
          <div className="space-x-4">
            <Link to="/" className="text-blue-500 hover:underline">Home</Link>
            <Link to="/online" className="text-blue-500 hover:underline">Quick Game</Link>
            <Link to="/offline" className="text-blue-500 hover:underline">Offline Game</Link>
          </div>
        </nav>

        <main className="flex-1 p-6">
          <Routes>
            <Route path="/" element={<MainPage />} />
            <Route path="/online" element={<OnlineGamePage />} />
            <Route path="/offline" element={<OfflineGamePage />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}
