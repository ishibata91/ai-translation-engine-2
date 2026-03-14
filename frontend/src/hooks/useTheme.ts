import {useEffect} from 'react';
import {useUIStore} from '../store/uiStore';

const THEME_STORAGE_KEY = 'ai_translation_engine_theme_v2';

/**
 * UI ストアと localStorage を同期し、利用可能なテーマ一覧を返す。
 */
export function useTheme() {
    const { theme, setTheme } = useUIStore();

    // Load theme from local storage on initial mount
    useEffect(() => {
        const savedTheme = localStorage.getItem(THEME_STORAGE_KEY);
        if (savedTheme) {
            setTheme(savedTheme);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    // Sync theme with HTML attribute and local storage when it changes
    useEffect(() => {
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem(THEME_STORAGE_KEY, theme);
    }, [theme]);

    const availableThemes = [
        "light", "dark"
    ];

    return { theme, setTheme, availableThemes };
}
