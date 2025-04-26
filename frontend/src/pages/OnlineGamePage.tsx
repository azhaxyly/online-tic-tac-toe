import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

export default function OnlineGamePage() {
    const navigate = useNavigate();
    const [ws, setWs] = useState<WebSocket | null>(null);
    const [status, setStatus] = useState('Searching for an opponent...');
    const [board, setBoard] = useState<string[]>(Array(9).fill(''));
    const [mySymbol, setMySymbol] = useState('');
    const [opponent, setOpponent] = useState('');
    const [gameResult, setGameResult] = useState('');
    const [isGameStarted, setIsGameStarted] = useState(false);

    useEffect(() => {
        const socket = new WebSocket(`ws://${location.host}/ws`);
        setWs(socket);

        socket.onopen = () => {
            socket.send(JSON.stringify({ type: 'find_match' }));
        };

        socket.onmessage = (event) => {
            const msg = JSON.parse(event.data);

            switch (msg.type) {
                case 'match_found':
                    setStatus(`Matched against ${msg.opponent}`);
                    setMySymbol(msg.symbol);
                    setOpponent(msg.opponent);
                    setIsGameStarted(true);
                    break;
                case 'move_made':
                    updateBoard(msg.cell, msg.by);
                    break;
                case 'game_over':
                    setGameResult(`Game Over: ${msg.result}`);
                    setIsGameStarted(false);
                    break;
                case 'opponent_left':
                    setStatus('Opponent left the game.');
                    setIsGameStarted(false);
                    break;
                case 'match_cancelled':
                    navigate('/');
                    break;
                default:
                    break;
            }
        };

        socket.onclose = () => {
            setStatus('Connection closed.');
        };

        return () => {
            socket.close();
        };
    }, [navigate]);

    function updateBoard(cell: number, symbol: string) {
        setBoard(prev => {
            const newBoard = [...prev];
            newBoard[cell] = symbol;
            return newBoard;
        });
    }

    function handleCellClick(index: number) {
        if (!isGameStarted || board[index] || !ws) return;
        ws.send(JSON.stringify({ type: 'move', cell: index }));
    }

    function handleCancel() {
        if (ws) {
            ws.send(JSON.stringify({ type: 'cancel_match' }));
        }
        navigate('/');
    }

    function handlePlayAgain() {
        if (ws) {
            ws.send(JSON.stringify({ type: 'play_again' }));
        }
        setBoard(Array(9).fill(''));
        setGameResult('');
    }

    return (
        <div className="flex flex-col items-center mt-10 space-y-6">
            <div className="text-2xl font-bold">{status}</div>

            {isGameStarted && (
                <>
                    <div className="text-lg">
                        You are playing as <b>{mySymbol}</b> against <b>{opponent}</b>
                    </div>
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

                    {gameResult && (
                        <div className="mt-4 text-xl font-semibold">{gameResult}</div>
                    )}

                    <div className="flex space-x-4 mt-4">
                        <button
                            onClick={handlePlayAgain}
                            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                        >
                            Play Again
                        </button>
                        <button
                            onClick={handleCancel}
                            className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600"
                        >
                            Back to Main
                        </button>
                    </div>
                </>
            )}
        </div>
    );
}
