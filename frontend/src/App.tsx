import { useState } from 'react'

// Wailsランタイムが window.go に注入するバインディングを型付きでラップ
// wailsjs/ ディレクトリは wails dev 実行後に自動生成されるが、
// ビルド時の依存を避けるため window.go を直接使用する
declare global {
    interface Window {
        go?: {
            main?: {
                App?: {
                    Greet: (name: string) => Promise<string>
                }
            }
        }
    }
}

const Greet = (name: string): Promise<string> => {
    if (window.go?.main?.App?.Greet) {
        return window.go.main.App.Greet(name)
    }
    // Wailsランタイム外（通常のブラウザ）でのフォールバック
    return Promise.resolve(`Hello, ${name}! (Wails runtime not available)`)
}

function App() {
    const [name, setName] = useState('')
    const [result, setResult] = useState('')

    const handleGreet = async () => {
        try {
            const response = await Greet(name)
            setResult(response)
        } catch (err) {
            console.error('Greet failed:', err)
            setResult('Error: Could not connect to backend')
        }
    }

    return (
        <div style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: '100vh',
            fontFamily: 'sans-serif',
            background: '#1b2636',
            color: '#fff',
        }}>
            <h1 style={{ fontSize: '2rem', marginBottom: '2rem' }}>
                AI Translation Engine
            </h1>
            <div style={{
                background: '#2d3a4a',
                padding: '2rem',
                borderRadius: '8px',
                minWidth: '360px',
                display: 'flex',
                flexDirection: 'column',
                gap: '1rem',
            }}>
                <label style={{ fontSize: '0.9rem', color: '#aaa' }}>名前:</label>
                <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleGreet()}
                    placeholder="名前を入力..."
                    style={{
                        padding: '0.5rem 0.75rem',
                        borderRadius: '4px',
                        border: '1px solid #3d4e60',
                        background: '#1b2636',
                        color: '#fff',
                        fontSize: '1rem',
                    }}
                />
                <button
                    onClick={handleGreet}
                    style={{
                        padding: '0.6rem 1.2rem',
                        borderRadius: '4px',
                        border: 'none',
                        background: '#4c72b0',
                        color: '#fff',
                        fontSize: '1rem',
                        cursor: 'pointer',
                    }}
                >
                    Greet
                </button>
                {result && (
                    <div style={{
                        padding: '0.75rem',
                        background: '#1b2636',
                        borderRadius: '4px',
                        marginTop: '0.5rem',
                        borderLeft: '3px solid #4c72b0',
                    }}>
                        <span style={{ color: '#aaa', fontSize: '0.85rem' }}>結果: </span>
                        <span>{result}</span>
                    </div>
                )}
            </div>
        </div>
    )
}

export default App
