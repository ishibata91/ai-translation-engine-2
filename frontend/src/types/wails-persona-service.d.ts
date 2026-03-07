declare module '*wailsjs/go/persona/Service' {
  export function ListNPCs(): Promise<unknown[]>;
  export function ListDialoguesBySpeaker(speakerID: string): Promise<unknown[]>;
}
