export interface SkillFile {
  id: string
  name: string
  path: string // e.g. "SKILL.md", "system-prompt.md", "references/profil-perusahaan.md"
  filePath?: string
  content: string
  isReference?: boolean
  isGlobal?: boolean
  updatedAt?: string
}

export interface Skill {
  id: string
  slug: string
  name: string
  description: string
  enabled: boolean
  isEnabled?: boolean
  files: SkillFile[]
  createdAt?: string
  updatedAt?: string
}

export type ViewMode = 'view' | 'edit'
