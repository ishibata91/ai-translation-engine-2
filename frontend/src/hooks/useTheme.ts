import { useEffect } from 'react';
import { useUIStore } from '../store/uiStore';

const THEME_STORAGE_KEY = 'ai_translation_engine_theme';

export function useTheme() {
    const { theme, setTheme } = useUIStore();

    // Load theme from local storage on initial mount
    useEffect(() => {
        const savedTheme = localStorage.getItem(THEME_STORAGE_KEY);
        if (savedTheme && savedTheme !== theme) {
            setTheme(savedTheme);
        }
    }, [setTheme]);

    // Sync theme with HTML attribute and local storage when it changes
    useEffect(() => {
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem(THEME_STORAGE_KEY, theme);
    }, [theme]);

    const availableThemes = [
        "light", "dark", "cupcake", "bumblebee", "emerald", "corporate",
        "synthwave", "retro", "cyberpunk", "valentine", "halloween",
        "garden", "forest", "aqua", "lofi", "pastel", "fantasy", "wireframe",
        "black", "luxury", "dracula", "cmyk", "autumn", "business", "acid",
        "lemonade", "night", "coffee", "winter", "dim", "nord", "sunset"
    ];

    return { theme, setTheme, availableThemes };
}
