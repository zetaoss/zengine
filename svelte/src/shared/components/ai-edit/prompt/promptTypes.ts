export interface TemplateVariable {
  name: string
}

export interface PromptPart {
  text: string
  type: 'plain' | 'inline' | 'block' | 'code'
  lang?: string
  source?: 'preset' | 'custom'
  ordinal?: number
  label?: string
  templateText?: string
  templateStart?: number
  templateEnd?: number
}
