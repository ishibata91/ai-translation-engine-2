export type NpcStatus = '完了' | '生成中' | '抽出完了' | 'エラー';

export interface Dialogue {
    recordType: string;
    editorId: string;
    source: string;
    translation: string;
}

export interface NpcRow {
    formId: string;
    name: string;
    dialogueCount: number;
    status: NpcStatus;
    updatedAt: string;
    promptHistory: string[];
    rawResponse: string;
    dialogues: Dialogue[];
}

export const STATUS_BADGE: Record<NpcStatus, string> = {
    '完了': 'badge-success',
    '生成中': 'badge-info',
    '抽出完了': 'badge-ghost',
    'エラー': 'badge-error',
};
