/**
 * Prompt loader utility
 * 
 * Dynamically imports prompt files at build time using Vite's glob import.
 * This allows prompts to be stored in .txt files and easily modified without code changes.
 */

// Import all .txt files from the prompts directory
const promptModules = import.meta.glob('./*.txt', { 
  query: '?raw',
  import: 'default',
  eager: true 
}) as Record<string, string>;

export const PromptType = {
  SYSTEM: 'system'
} as const;

export type PromptType = typeof PromptType[keyof typeof PromptType];

/**
 * Get a specific prompt by type
 */
export function getPrompt(type: PromptType): string {
  const filename = `./${type}.txt`;
  const prompt = promptModules[filename];
  
  if (!prompt) {
    throw new Error(`Prompt not found: ${type}`);
  }
  
  return prompt;
}

/**
 * Get all available prompts
 */
export function getAllPrompts(): Record<PromptType, string> {
  return {
    [PromptType.SYSTEM]: getPrompt(PromptType.SYSTEM),
  };
}

/**
 * List all available prompt types
 */
export function getAvailablePromptTypes(): PromptType[] {
  return Object.values(PromptType);
}

export default {
  getPrompt,
  getAllPrompts,
  getAvailablePromptTypes,
  PromptType,
};
