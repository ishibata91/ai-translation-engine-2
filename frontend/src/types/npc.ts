export type NpcStatus = 'draft' | 'generated';

export interface Dialogue {
    recordType: string;
    editorId: string;
    source: string;
}

export interface NpcRow {
    id: string;
    personaId: number;
    formId: string;
    sourcePlugin: string;
    name: string;
    race: string;
    sex: string;
    voiceType: string;
    dialogueCount: number;
    status: NpcStatus;
    updatedAt: string;
    personaText: string;
    generationRequest: string;
    dialogues: Dialogue[];
}

export const NPC_STATUS_LABEL: Record<NpcStatus, string> = {
    draft: '下書き',
    generated: '生成済み',
};

export const STATUS_BADGE: Record<NpcStatus, string> = {
    draft: 'badge-ghost',
    generated: 'badge-success',
};
