'use client'

import { useEffect, useState } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from '@tanstack/react-router'
import {
  CreateKnowledgeRequest,
  UpdateKnowledgeRequest,
} from '@/gen/v1/knowledge_pb'
import MarkdownPreview from '@uiw/react-markdown-preview'
import MDEditor from '@uiw/react-md-editor'
import { AlertCircle, ArrowLeft, Info, PencilLine, X } from 'lucide-react'
import { toast } from 'sonner'
import { useTheme } from '@/context/theme-provider'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import {
  useCreateKnowledgeMutation,
  useGetKnowledgeQuery,
  useUpdateKnowledgeMutation,
} from './api/use-knowledge'

const formSchema = z.object({
  title: z.string().min(1, 'Title is required.'),
  category: z.string().optional(),
  tags: z.string().optional(),
  content: z.string().optional(),
  embedToLlm: z.boolean(),
})

function splitTags(raw: string): string[] {
  return raw
    .split(',')
    .map((t) => t.trim())
    .filter(Boolean)
}

interface KnowledgeEditorPageProps {
  mode: 'create' | 'edit'
  id?: string
}

export function KnowledgeEditorPage({ mode, id }: KnowledgeEditorPageProps) {
  const isEdit = mode === 'edit'
  const navigate = useNavigate()
  const { theme } = useTheme()
  const colorMode = theme === 'dark' ? 'dark' : 'light'

  const getQuery = useGetKnowledgeQuery(isEdit ? id : undefined)
  const createMutation = useCreateKnowledgeMutation()
  const updateMutation = useUpdateKnowledgeMutation()

  // GitHub-like: default menampilkan preview hasil render; mode edit hanya
  // muncul saat tombol Edit diklik, dan bisa ditutup kembali ke preview.
  // Tidak ada split view preview+edit bersamaan.
  const [editorMode, setEditorMode] = useState<'preview' | 'edit'>('preview')

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      title: '',
      category: '',
      tags: '',
      content: '',
      embedToLlm: false,
    },
  })

  useEffect(() => {
    if (isEdit) {
      if (getQuery.data) {
        form.reset({
          title: getQuery.data.title,
          category: getQuery.data.category,
          tags: (getQuery.data.tags ?? []).join(', '),
          content: getQuery.data.content,
          embedToLlm: getQuery.data.embedToLlm,
        })
      }
    } else {
      form.reset({
        title: '',
        category: '',
        tags: '',
        content: '',
        embedToLlm: false,
      })
    }
  }, [isEdit, getQuery.data, form])

  async function onSubmit(values: z.infer<typeof formSchema>) {
    try {
      const tags = splitTags(values.tags ?? '')
      const content = values.content ?? ''
      const category = values.category ?? ''

      if (isEdit && id) {
        await updateMutation.mutateAsync(
          new UpdateKnowledgeRequest({
            id,
            title: values.title,
            content,
            category,
            tags,
            embedToLlm: values.embedToLlm,
          })
        )
        toast.success('Document updated successfully')
      } else {
        await createMutation.mutateAsync(
          new CreateKnowledgeRequest({
            title: values.title,
            content,
            category,
            tags,
            embedToLlm: values.embedToLlm,
          })
        )
        toast.success('Document created successfully')
      }
      navigate({ to: '/knowledge' })
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to save document'
      toast.error(errorMessage)
    }
  }

  const isSaving = createMutation.isPending || updateMutation.isPending

  if (isEdit && getQuery.isPending) {
    return (
      <>
        <Header fixed>
          <h1 className='text-sm font-medium'>Loading document...</h1>
        </Header>
        <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
          <p className='text-sm text-muted-foreground'>
            Memuat dokumen knowledge...
          </p>
        </Main>
      </>
    )
  }

  if (isEdit && (getQuery.isError || !getQuery.data)) {
    return (
      <>
        <Header fixed>
          <Button
            variant='ghost'
            size='sm'
            className='gap-1'
            onClick={() => navigate({ to: '/knowledge' })}
          >
            <ArrowLeft className='h-4 w-4' />
            Back
          </Button>
        </Header>
        <Main className='flex flex-1 flex-col items-center justify-center gap-3'>
          <h1 className='text-lg font-semibold'>Document not found</h1>
          <p className='text-sm text-muted-foreground'>
            Dokumen yang Anda cari sudah dihapus atau tidak tersedia.
          </p>
          <Button
            variant='outline'
            onClick={() => navigate({ to: '/knowledge' })}
          >
            Back to Knowledge Base
          </Button>
        </Main>
      </>
    )
  }

  const embedFailed = isEdit && getQuery.data?.embedStatus === 'failed'

  return (
    <>
      <Header fixed>
        <Button
          variant='ghost'
          size='sm'
          className='gap-1'
          onClick={() => navigate({ to: '/knowledge' })}
        >
          <ArrowLeft className='h-4 w-4' />
          Back
        </Button>
        <h1 className='truncate text-sm font-medium'>
          {isEdit ? 'Edit Document' : 'New Document'}
        </h1>
        <div className='ms-auto flex items-center gap-2'>
          <Button
            variant='outline'
            size='sm'
            onClick={() => navigate({ to: '/knowledge' })}
          >
            Cancel
          </Button>
          <Button
            type='submit'
            form='knowledge-form'
            size='sm'
            disabled={isSaving}
          >
            {isSaving
              ? 'Saving...'
              : isEdit
                ? 'Save Changes'
                : 'Create Document'}
          </Button>
        </div>
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <Form {...form}>
          <form
            id='knowledge-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className='mx-auto w-full max-w-4xl space-y-6'
          >
            <FormField
              control={form.control}
              name='title'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Title</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='Prosedur Reset Router Mikrotik'
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
              <FormField
                control={form.control}
                name='category'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Category (optional)</FormLabel>
                    <FormControl>
                      <Input placeholder='umum' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='tags'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Tags (comma separated)</FormLabel>
                    <FormControl>
                      <Input
                        placeholder='mikrotik, reset, troubleshooting'
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name='content'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Content (Markdown)</FormLabel>
                  <FormControl>
                    <div
                      data-color-mode={colorMode}
                      className='overflow-hidden rounded-md border'
                    >
                      <div className='flex items-center justify-between border-b bg-muted/40 px-3 py-1.5'>
                        <span className='text-xs font-medium text-muted-foreground'>
                          {editorMode === 'preview' ? 'Preview' : 'Editing'}
                        </span>
                        {editorMode === 'preview' ? (
                          <Button
                            type='button'
                            variant='ghost'
                            size='sm'
                            className='h-7 gap-1 text-xs'
                            onClick={() => setEditorMode('edit')}
                          >
                            <PencilLine className='h-3.5 w-3.5' />
                            Edit
                          </Button>
                        ) : (
                          <Button
                            type='button'
                            variant='ghost'
                            size='sm'
                            className='h-7 gap-1 text-xs'
                            onClick={() => setEditorMode('preview')}
                          >
                            <X className='h-3.5 w-3.5' />
                            Close edit
                          </Button>
                        )}
                      </div>
                      {editorMode === 'preview' ? (
                        field.value ? (
                          <MarkdownPreview
                            source={field.value}
                            style={{ padding: '20px 24px' }}
                          />
                        ) : (
                          <p className='px-5 py-6 text-sm text-muted-foreground'>
                            Belum ada konten. Klik <strong>Edit</strong> untuk
                            mulai menulis dalam format markdown.
                          </p>
                        )
                      ) : (
                        <MDEditor
                          value={field.value}
                          onChange={(value) => field.onChange(value ?? '')}
                          height={480}
                          preview='edit'
                          extraCommands={[]}
                          visibleDragbar={false}
                          textareaProps={{
                            placeholder:
                              'Tulis isi dokumen dalam format markdown...',
                          }}
                        />
                      )}
                    </div>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {embedFailed && (
              <div className='flex items-start gap-2 rounded-md border border-red-200 bg-red-50 p-3 text-xs text-red-700 dark:border-red-900 dark:bg-red-950 dark:text-red-300'>
                <AlertCircle className='mt-0.5 h-3.5 w-3.5 shrink-0' />
                <span>
                  Embed ke AnythingLLM sebelumnya gagal. Simpan dokumen ini
                  untuk mencoba sinkronisasi lagi, atau gunakan menu &quot;Retry
                  Embed&quot; di tabel.
                </span>
              </div>
            )}

            <FormField
              control={form.control}
              name='embedToLlm'
              render={({ field }) => (
                <FormItem className='flex flex-row items-center justify-between rounded-lg border p-3 shadow-xs'>
                  <div className='space-y-0.5'>
                    <FormLabel>Embed to AnythingLLM</FormLabel>
                    <p className='flex items-center gap-1 text-xs text-muted-foreground'>
                      <Info className='h-3 w-3' />
                      Aktifkan agar dokumen ikut di-embed ke vector store dan
                      bisa dipakai bot WhatsApp. Nonaktif = dokumen hanya di
                      dashboard admin.
                    </p>
                  </div>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </FormItem>
              )}
            />
          </form>
        </Form>
      </Main>
    </>
  )
}
