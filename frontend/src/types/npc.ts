interface Dialogue {
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
    updatedAt: string;
    personaText: string;
    generationRequest: string;
    dialogues: Dialogue[];
}
