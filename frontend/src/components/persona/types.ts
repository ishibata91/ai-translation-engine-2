export type PersonaListStateTone = 'neutral' | 'info' | 'warning' | 'success' | 'error';

interface PersonaListStateBadge {
    label: string;
    tone?: PersonaListStateTone;
}

export interface PersonaListRow {
    id: string;
    formId: string;
    sourcePlugin: string;
    npcName: string;
    editorId?: string;
    updatedAt?: string;
    stateBadge?: PersonaListStateBadge | null;
}

export interface PersonaListPager {
    page: number;
    totalPages: number;
    onPrevPage?: () => void;
    onNextPage?: () => void;
    disablePrev?: boolean;
    disableNext?: boolean;
}
