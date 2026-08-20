import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { botClient } from '@/lib/api-client'
import { toast } from 'sonner'
import type { Skill } from '../types'

export const skillKeys = {
  all: ['skills'] as const,
  lists: () => [...skillKeys.all, 'list'] as const,
  detail: (slug: string) => [...skillKeys.all, 'detail', slug] as const,
  globalPrompt: () => [...skillKeys.all, 'global-prompt'] as const,
}

export function useSkills() {
  return useQuery({
    queryKey: skillKeys.lists(),
    queryFn: async () => {
      const resp = await botClient.listSkills({})
      const mapped: Skill[] = resp.skills.map((s) => ({
        id: String(s.id),
        slug: s.slug,
        name: s.name,
        description: s.description,
        enabled: s.isEnabled,
        isEnabled: s.isEnabled,
        files: s.files.map((f) => ({
          id: String(f.id),
          skillId: String(f.skillId),
          name: f.name,
          path: f.filePath,
          filePath: f.filePath,
          content: f.content,
          isReference: f.isReference,
          updatedAt: f.updatedAt,
        })),
        createdAt: s.createdAt,
        updatedAt: s.updatedAt,
      }))
      return mapped
    },
  })
}

export function useGlobalPrompt() {
  return useQuery({
    queryKey: skillKeys.globalPrompt(),
    queryFn: async () => {
      const resp = await botClient.getGlobalPrompt({})
      return resp.content
    },
  })
}

export function useSaveGlobalPrompt() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (content: string) => {
      return await botClient.saveGlobalPrompt({ content })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: skillKeys.globalPrompt() })
      toast.success('Global System Prompt berhasil disimpan ke DB & Disk')
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
      return await botClient.createSkill({ slug, name, description })
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
      isReference,
    }: {
      slug: string
      filePath: string
      content: string
      isReference: boolean
    }) => {
      return await botClient.saveSkillFile({
        slug,
        filePath,
        content,
        isReference,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: skillKeys.lists() })
      toast.success('Berkas berhasil disimpan ke Database & Disk')
    },
    onError: (err: Error) => {
      toast.error(`Gagal menyimpan berkas: ${err.message}`)
    },
  })
}

export function useDeleteSkill() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, slug }: { id: number; slug: string }) => {
      return await botClient.deleteSkill({ id, slug })
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
      fileId,
      filePath,
    }: {
      slug: string
      fileId: number
      filePath: string
    }) => {
      return await botClient.deleteSkillFile({
        slug,
        fileId,
        filePath,
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
      return await botClient.toggleSkill({ slug, enabled })
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

export function useSyncSkillsFromDisk() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      return await botClient.syncSkillsFromDisk({})
    },
    onSuccess: (resp) => {
      queryClient.invalidateQueries({ queryKey: skillKeys.lists() })
      queryClient.invalidateQueries({ queryKey: skillKeys.globalPrompt() })
      toast.success(resp.message || 'Berhasil sinkronisasi dari disk')
    },
    onError: (err: Error) => {
      toast.error(`Gagal sinkronisasi dari disk: ${err.message}`)
    },
  })
}
