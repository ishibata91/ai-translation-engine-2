export type TaskStatus = 'pending' | 'running' | 'paused' | 'completed' | 'request_generated' | 'failed' | 'cancelled';

export type TaskType = 'dictionary_build' | 'persona_extraction' | 'translation_project';

export interface FrontendTask {
    id: string;
    name: string;
    type: TaskType;
    status: TaskStatus;
    phase: string;
    progress: number;
    error_msg?: string;
    metadata: Record<string, any>;
    created_at: string;
    updated_at: string;
}

export interface PhaseCompletedEvent {
    taskId: string;
    phase: string;
    summary: any;
}
