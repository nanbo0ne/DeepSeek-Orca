export interface ComposerDraftStateInput {
  text: string;
  attachmentCount: number;
  workspaceReferenceCount: number;
  sessionReferenceCount: number;
  pendingPasteCount: number;
}

export interface ComposerDraftState {
  hasDraftContent: boolean;
  hasSendableContent: boolean;
}

export function composerDraftState(input: ComposerDraftStateInput): ComposerDraftState {
  const hasSendableContent = Boolean(
    input.text.trim() ||
      input.attachmentCount > 0 ||
      input.workspaceReferenceCount > 0 ||
      input.sessionReferenceCount > 0,
  );
  return {
    hasDraftContent: hasSendableContent || input.pendingPasteCount > 0,
    hasSendableContent,
  };
}
