import React from 'react';

interface PromptSettingCardProps {
    title: string;
    description: string;
    value: string;
    onChange?: (value: string) => void;
    readOnly?: boolean;
    badgeLabel: string;
    footerText?: string;
}

const PromptSettingCard: React.FC<PromptSettingCardProps> = ({
    title,
    description,
    value,
    onChange,
    readOnly = false,
    badgeLabel,
    footerText,
}) => (
    <div className="card bg-base-100 border border-base-200 shadow-sm h-full">
        <div className="card-body gap-3">
            <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
                <div className="space-y-1">
                    <h2 className="card-title text-base">{title}</h2>
                    <p className="text-sm text-base-content/70">{description}</p>
                </div>
                <span className={`badge badge-sm shrink-0 whitespace-nowrap ${readOnly ? 'badge-ghost' : 'badge-primary badge-outline'}`}>
                    {badgeLabel}
                </span>
            </div>
            <textarea
                className="textarea textarea-bordered min-h-56 w-full font-mono text-xs leading-6"
                value={value}
                readOnly={readOnly}
                onChange={readOnly || !onChange ? undefined : (event) => onChange(event.target.value)}
            />
            <span className="text-xs text-base-content/60">
                {footerText ?? (readOnly ? 'このカードは送信時の system prompt をそのまま表示します。' : '変更内容は自動保存され、次回表示時にも復元されます。')}
            </span>
        </div>
    </div>
);

export default PromptSettingCard;
