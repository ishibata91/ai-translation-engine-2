/**
 * ペルソナ生成進捗イベントの状態値。
 */
type PersonaProgressStatus = 'IN_PROGRESS' | 'COMPLETED' | 'FAILED';

/**
 * ペルソナ生成進捗イベントの payload。
 */
export interface PersonaProgressEvent {
    CorrelationID: string;
    TaskID?: string;
    TaskType?: string;
    Phase?: string;
    Current?: number;
    Total: number;
    Completed: number;
    Failed: number;
    Status: PersonaProgressStatus;
    Message: string;
}

/**
 * ペルソナ一覧へ整形する前の NPC レコード。
 */
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
    persona_text?: string;
    PersonaText?: string;
    generation_request?: string;
    GenerationRequest?: string;
    updated_at?: string;
    UpdatedAt?: string;
}

/**
 * ペルソナ詳細へ整形する前の会話レコード。
 */
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
