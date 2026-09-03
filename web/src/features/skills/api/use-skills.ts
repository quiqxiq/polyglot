import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { skillClient } from '@/lib/api-client'
import { toast } from 'sonner'
import type { Skill, SkillFile } from '../types'

export const skillKeys = {
  all: ['skills'] as const,
  lists: () => [...skillKeys.all, 'list'] as const,
  detail: (slug: string) => [...skillKeys.all, 'detail', slug] as const,
  globalPrompt: () => [...skillKeys.all, 'global-prompt'] as const,
}

function formatSkillFrontmatter(s: {
  name: string
  description: string
  license?: string
  compatibility?: string
  allowedTools?: string
  metadata?: Record<string, string>
}) {
  let fm = `---\nname: ${s.name}\ndescription: '${(s.description || '').replace(/'/g, "''")}'\n`
  if (s.license) fm += `license: ${s.license}\n`
  if (s.compatibility) fm += `compatibility: ${s.compatibility}\n`
  if (s.allowedTools) fm += `allowed-tools: ${s.allowedTools}\n`
  if (s.metadata && Object.keys(s.metadata).length > 0) {
    fm += `metadata:\n`
    for (const [k, v] of Object.entries(s.metadata)) {
      fm += `  ${k}: ${v}\n`
    }
  }
  fm += `---\n\n`
  return fm
}

export function useSkills() {
  return useQuery({
    queryKey: skillKeys.lists(),
    queryFn: async () => {
      const resp = await skillClient.listSkills({})
      const mapped: Skill[] = await Promise.all(
        resp.skills.map(async (s) => {
          let resourceFiles: SkillFile[] = []
          try {
            const resResp = await skillClient.listResources({ skillId: s.id })
            resourceFiles = (resResp.resources || []).map((r) => ({
              id: `${s.id}-${r.path}`,
              skillId: s.id,
              name: r.name,
              path: r.path,
              filePath: r.path,
              content: '',
              isReference: r.type === 'reference' || r.path.startsWith('references/'),
              updatedAt: r.modified,
            }))
          } catch {
            // Ignore error if skill has no resources
          }

          const rawContent = s.content || ''
          const fullContent = rawContent.startsWith('---')
            ? rawContent
            : `${formatSkillFrontmatter(s)}${rawContent}`

          const mainFile: SkillFile = {
            id: `${s.id}-main`,
            skillId: s.id,
            name: 'SKILL.md',
            path: 'SKILL.md',
            filePath: 'SKILL.md',
            content: fullContent,
            isReference: false,
            updatedAt: s.updatedAt,
          }

          return {
            id: s.id,
            slug: s.name,
            name: s.name,
            description: s.description,
            enabled: true,
            isEnabled: true,
            files: [mainFile, ...resourceFiles],
            createdAt: s.createdAt,
            updatedAt: s.updatedAt,
          }
        })
      )
      return mapped
    },
  })
}

export function useGlobalPrompt() {
  return useQuery({
    queryKey: skillKeys.globalPrompt(),
    queryFn: async () => {
      const resp = await skillClient.getGlobalPrompt({})
      return resp.content
    },
  })
}

export function useSaveGlobalPrompt() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (content: string) => {
      return await skillClient.saveGlobalPrompt({ content })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: skillKeys.globalPrompt() })
      toast.success('Global System Prompt berhasil disimpan')
    },
    onError: (err: Error) => {
      toast.error(`Gagal menyimpan global prompt: ${err.message}`)
    },
  })
}

export function useCreateSkill() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      slug,
      name,
      description,
    }: {
      slug: string
      name: string
      description: string
    }) => {
      const skillName = (slug || name).toLowerCase().replace(/\s+/g, '-')
      return await skillClient.createSkill({
        name: skillName,
        description,
        content: `# ${name}\n\n${description}\n`,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: skillKeys.lists() })
      toast.success('Skill baru berhasil dibuat')
    },
    onError: (err: Error) => {
      toast.error(`Gagal membuat skill: ${err.message}`)
    },
  })
}

export function useSaveSkillFile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      slug,
      filePath,
      content,
    }: {
      slug: string
      filePath: string
      content: string
      isReference?: boolean
    }) => {
      if (filePath === 'SKILL.md' || !filePath) {
        return await skillClient.updateSkill({
          id: slug,
          name: slug,
          content: content,
        })
      }
      return await skillClient.saveResource({
        skillId: slug,
        path: filePath,
        data: new TextEncoder().encode(content),
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: skillKeys.lists() })
      toast.success('Berkas berhasil disimpan')
    },
    onError: (err: Error) => {
      toast.error(`Gagal menyimpan berkas: ${err.message}`)
    },
  })
}

export function useDeleteSkill() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, slug }: { id: number | string; slug: string }) => {
      const targetId = slug || String(id)
      return await skillClient.deleteSkill({ id: targetId })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: skillKeys.lists() })
      toast.success('Skill berhasil dihapus')
    },
    onError: (err: Error) => {
      toast.error(`Gagal menghapus skill: ${err.message}`)
    },
  })
}

export function useDeleteSkillFile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      slug,
      filePath,
    }: {
      slug: string
      fileId?: number | string
      filePath: string
    }) => {
      return await skillClient.deleteResource({
        skillId: slug,
        path: filePath,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: skillKeys.lists() })
      toast.success('Berkas berhasil dihapus')
    },
    onError: (err: Error) => {
      toast.error(`Gagal menghapus berkas: ${err.message}`)
    },
  })
}

export function useToggleSkill() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ slug, enabled }: { slug: string; enabled: boolean }) => {
      return await skillClient.toggleSkill({ id: slug, enabled })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: skillKeys.lists() })
      toast.success('Status skill diperbarui')
    },
    onError: (err: Error) => {
      toast.error(`Gagal mengubah status skill: ${err.message}`)
    },
  })
}

