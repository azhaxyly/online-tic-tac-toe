import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

export default function OfflineGamePage() {
    const navigate = useNavigate();
    const [board, setBoard] = useState<string[]>(Array(9).fill(''));
    const [turn, setTurn] = useState<'X' | 'O'>('X');
    const [winner, setWinner] = useState('');
    const [stats, setStats] = useState({ X: 0, O: 0, draws: 0 });

    function checkWinner(board: string[]): string {
        const lines = [
            [0, 1, 2], [3, 4, 5], [6, 7, 8],
            [0, 3, 6], [1, 4, 7], [2, 5, 8],
            [0, 4, 8], [2, 4, 6]
        ];

        for (const [a, b, c] of lines) {
            if (board[a] && board[a] === board[b] && board[b] === board[c]) {
                return board[a];
            }
        }

        if (board.every(cell => cell)) return 'Draw';
        return '';
    }

    function handleCellClick(index: number) {
        if (board[index] || winner) return;

        const newBoard = [...board];
        newBoard[index] = turn;
        const newWinner = checkWinner(newBoard);

        setBoard(newBoard);

        if (newWinner) {
            setWinner(newWinner);
            if (newWinner === 'Draw') {
                setStats(prev => ({ ...prev, draws: prev.draws + 1 }));
            } else {
                setStats(prev => ({ ...prev, [newWinner]: prev[newWinner as 'X' | 'O'] + 1 }));
            }

            setTimeout(() => resetGame(), 1500); // Автоматический рестарт через 1.5 секунды
        } else {
            setTurn(prev => (prev === 'X' ? 'O' : 'X'));
        }
    }

    function resetGame() {
        setBoard(Array(9).fill(''));
        setTurn('X');
        setWinner('');
    }

    function backToMenu() {
        navigate('/');
    }

    return (
        <div className="flex flex-col items-center mt-10 space-y-6">
            <div className="text-2xl font-bold">Offline Game</div>

            <div className="grid grid-cols-3 gap-2">
                {board.map((cell, idx) => (
                    <div
                        key={idx}
                        onClick={() => handleCellClick(idx)}
                        className="w-20 h-20 bg-white border flex items-center justify-center text-3xl cursor-pointer hover:bg-gray-100"
                    >
                        {cell}
                    </div>
                ))}
            </div>

            {winner && (
                <div className="mt-4 text-xl font-semibold">
                    {winner === 'Draw' ? 'It\'s a draw!' : `${winner} wins!`}
                </div>
            )}

            <div className="flex space-x-4 mt-4">
                <button
                    onClick={backToMenu}
                    className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600"
                >
                    Back to Main Menu
                </button>
            </div>

            <div className="mt-6 text-md">
                <div>Statistics:</div>
                <div>X wins: {stats.X}</div>
                <div>O wins: {stats.O}</div>
                <div>Draws: {stats.draws}</div>
            </div>
        </div>
    );
}
