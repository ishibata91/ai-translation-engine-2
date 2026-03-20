import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'
import {ensureWailsBridge} from './bootstrap/ensureWailsBridge'

ensureWailsBridge()

ReactDOM.createRoot(document.getElementById('root')!).render(
    <React.StrictMode>
        <App />
    </React.StrictMode>,
)
