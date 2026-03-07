export type PersonaProgressStatus = 'IN_PROGRESS' | 'COMPLETED' | 'FAILED';

export interface PersonaProgressEvent {
    CorrelationID: string;
    Total: number;
    Completed: number;
    Failed: number;
    Status: PersonaProgressStatus;
    Message: string;
}

export interface PersonaNPCRecord {
    persona_id?: number;
    PersonaID?: number;
    speaker_id?: string;
    SpeakerID?: string;
    source_plugin?: string;
    SourcePlugin?: string;
    npc_name?: string;
    NPCName?: string;
    race?: string;
    Race?: string;
    sex?: string;
    Sex?: string;
    voice_type?: string;
    VoiceType?: string;
    dialogue_count?: number;
    DialogueCount?: number;
    persona_text?: string;
    PersonaText?: string;
    generation_request?: string;
    GenerationRequest?: string;
    status?: string;
    Status?: string;
    updated_at?: string;
    UpdatedAt?: string;
}

export interface PersonaDialogueRecord {
    persona_id?: number;
    PersonaID?: number;
    record_type?: string;
    RecordType?: string;
    editor_id?: string;
    EditorID?: string;
    source_text?: string;
    SourceText?: string;
}
