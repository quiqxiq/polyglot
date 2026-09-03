import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import {
  Cpu,
  Zap,
  DollarSign,
  Plus,
  Loader2,
  Trash2,
  Pencil,
  Edit3,
} from 'lucide-react'
import {
  useLLMConfigsQuery,
  useCreateLLMConfigMutation,
  useUpdateLLMConfigMutation,
  useActivateLLMConfigMutation,
  useTestLLMConfigMutation,
  useDeleteLLMConfigMutation,
} from '@/features/skills/api/use-llm-config'
import {
  CreateLLMConfigRequest,
  UpdateLLMConfigRequest,
  ActivateLLMConfigRequest,
  DeleteLLMConfigRequest,
  TestLLMConfigRequest,
} from '@/gen/v1/llm_pb'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { toast } from 'sonner'

interface LLMConfigDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

interface ModelPreset {
  id: string
  name: string
  inputPrice: number
  outputPrice: number
  note: string
}

const PROVIDER_PRESETS: Record<
  string,
  {
    name: string
    models: ModelPreset[]
  }
> = {
  gemini: {
    name: 'Google Gemini (Google AI Studio)',
    models: [
      {
        id: 'gemini-2.0-flash',
        name: 'gemini-2.0-flash (Rekomendasi Utama Google)',
        inputPrice: 0.1,
        outputPrice: 0.4,
        note: 'Sangat cepat (<0.7s), penalaran tinggi, dan hemat biaya',
      },
      {
        id: 'gemini-2.0-flash-lite',
        name: 'gemini-2.0-flash-lite (Ultra Hemat)',
        inputPrice: 0.075,
        outputPrice: 0.3,
        note: 'Ekonomis untuk traffic percakapan sangat tinggi',
      },
      {
        id: 'gemini-1.5-flash',
        name: 'gemini-1.5-flash',
        inputPrice: 0.075,
        outputPrice: 0.3,
        note: 'Model stabil dan handal untuk CS bot',
      },
      {
        id: 'gemini-1.5-pro',
        name: 'gemini-1.5-pro',
        inputPrice: 1.25,
        outputPrice: 5.0,
        note: 'Penalaran tingkat tinggi & konteks besar',
      },
    ],
  },
  groq: {
    name: 'Groq Cloud (Ultra Low Latency LPU)',
    models: [
      {
        id: 'qwen/qwen3.6-27b',
        name: 'qwen/qwen3.6-27b (Sangat Cerdas & Teruji)',
        inputPrice: 0.6,
        outputPrice: 3.0,
        note: 'Akurasi pemahaman konteks SOP tertinggi di Groq Cloud',
      },
      {
        id: 'openai/gpt-oss-120b',
        name: 'openai/gpt-oss-120b (Penalaran Mendalam)',
        inputPrice: 0.15,
        outputPrice: 0.6,
        note: 'Model penalaran instruksi tingkat enterprise',
      },
      {
        id: 'openai/gpt-oss-20b',
        name: 'openai/gpt-oss-20b (Cepat & Hemat)',
        inputPrice: 0.075,
        outputPrice: 0.3,
        note: 'Model tanggap kilat untuk FAQ dan percakapan rutin',
      },
      {
        id: 'llama-3.3-70b-versatile',
        name: 'llama-3.3-70b-versatile',
        inputPrice: 0.59,
        outputPrice: 0.79,
        note: 'Flagship Meta LLaMA 3.3 (jika diaktifkan di akun Groq)',
      },
      {
        id: 'llama-3.1-8b-instant',
        name: 'llama-3.1-8b-instant',
        inputPrice: 0.05,
        outputPrice: 0.08,
        note: 'Super ringan dan super murah',
      },
      {
        id: 'groq/compound',
        name: 'groq/compound (Groq Compound Router)',
        inputPrice: 0.5,
        outputPrice: 1.5,
        note: 'Orkestrasi model multi-domain otomatis',
      },
      {
        id: 'deepseek-r1-distill-llama-70b',
        name: 'deepseek-r1-distill-llama-70b',
        inputPrice: 0.75,
        outputPrice: 0.99,
        note: 'Penalaran langkah-demi-langkah (Step-by-step SOP)',
      },
    ],
  },
  openai: {
    name: 'OpenAI',
    models: [
      {
        id: 'gpt-4o-mini',
        name: 'gpt-4o-mini (Standar Industri CS)',
        inputPrice: 0.15,
        outputPrice: 0.6,
        note: 'Standar emas untuk bot pelayanan pelanggan',
      },
      {
        id: 'gpt-4o',
        name: 'gpt-4o (Flagship)',
        inputPrice: 2.5,
        outputPrice: 10.0,
        note: 'Kualitas respon tertinggi untuk kasus kompleks',
      },
      {
        id: 'o3-mini',
        name: 'o3-mini (Reasoning Model)',
        inputPrice: 1.1,
        outputPrice: 4.4,
        note: 'Model penalaran komputasi mendalam OpenAI',
      },
    ],
  },
  deepseek: {
    name: 'DeepSeek API',
    models: [
      {
        id: 'deepseek-chat',
        name: 'deepseek-chat (DeepSeek-V3)',
        inputPrice: 0.14,
        outputPrice: 0.28,
        note: 'Sangat cerdas dan biaya token paling terjangkau',
      },
      {
        id: 'deepseek-reasoner',
        name: 'deepseek-reasoner (DeepSeek-R1)',
        inputPrice: 0.55,
        outputPrice: 2.19,
        note: 'Model reasoning setara OpenAI o1',
      },
    ],
  },
  claude: {
    name: 'Anthropic Claude',
    models: [
      {
        id: 'claude-3-5-haiku-latest',
        name: 'claude-3-5-haiku (Cepat & Humanis)',
        inputPrice: 0.8,
        outputPrice: 4.0,
        note: 'Gaya bahasa percakapan sangat alami dan ramah',
      },
      {
        id: 'claude-3-5-sonnet-latest',
        name: 'claude-3-5-sonnet (Flagship Anthropic)',
        inputPrice: 3.0,
        outputPrice: 15.0,
        note: 'Pemahaman instruksi rumit terbaik di kelasnya',
      },
    ],
  },
  ollama: {
    name: 'Ollama (Lokal / Private Server)',
    models: [
      {
        id: 'qwen2.5:7b',
        name: 'qwen2.5:7b (Lokal On-Premise)',
        inputPrice: 0.0,
        outputPrice: 0.0,
        note: '100% Gratis & Mandiri di server lokal',
      },
      {
        id: 'llama3.2:3b',
        name: 'llama3.2:3b (Lokal Ringan)',
        inputPrice: 0.0,
        outputPrice: 0.0,
        note: 'Ringan untuk PC / server lokal dengan RAM terbatas',
      },
      {
        id: 'deepseek-r1:8b',
        name: 'deepseek-r1:8b (Lokal Reasoning)',
        inputPrice: 0.0,
        outputPrice: 0.0,
        note: 'Model penalaran offline mandiri',
      },
    ],
  },
  custom: {
    name: 'Custom (OpenAI-Compatible / OpenRouter / vLLM)',
    models: [
      {
        id: 'custom-model',
        name: 'Model Mandiri (Ketik Manual ID Model)',
        inputPrice: 0.5,
        outputPrice: 1.5,
        note: 'Gunakan dengan OpenRouter, Together AI, atau vLLM lokal',
      },
    ],
  },
}

export function LLMConfigDialog({ open, onOpenChange }: LLMConfigDialogProps) {
  const { data: configs, isLoading } = useLLMConfigsQuery()
  const createMutation = useCreateLLMConfigMutation()
  const updateMutation = useUpdateLLMConfigMutation()
  const activateMutation = useActivateLLMConfigMutation()
  const testMutation = useTestLLMConfigMutation()
  const deleteMutation = useDeleteLLMConfigMutation()

  const [isFormOpen, setIsFormOpen] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)

  const [provider, setProvider] = useState('groq')
  const [model, setModel] = useState('qwen/qwen3.6-27b')
  const [isCustomModel, setIsCustomModel] = useState(false)
  const [customModelInput, setCustomModelInput] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [maxTokens, setMaxTokens] = useState('1024')

  // State untuk modal konfirmasi hapus LLM
  const [configToDelete, setConfigToDelete] = useState<any | null>(null)

  const selectedPreset = PROVIDER_PRESETS[provider] || PROVIDER_PRESETS.groq
  const effectiveModelId = isCustomModel ? customModelInput.trim() : model
  const selectedModel =
    selectedPreset.models.find((m) => m.id === effectiveModelId) || {
      id: effectiveModelId,
      name: effectiveModelId,
      inputPrice: 0.5,
      outputPrice: 1.5,
      note: 'Model kustom pengguna',
    }

  const handleProviderChange = (newProv: string) => {
    setProvider(newProv)
    const preset = PROVIDER_PRESETS[newProv]
    if (preset && preset.models.length > 0) {
      setModel(preset.models[0].id)
      setIsCustomModel(false)
      setCustomModelInput('')
    }
  }

  const handleModelSelectChange = (val: string) => {
    if (val === '__custom__') {
      setIsCustomModel(true)
    } else {
      setIsCustomModel(false)
      setModel(val)
    }
  }

  const handleOpenAdd = () => {
    setEditingId(null)
    setProvider('groq')
    setModel('qwen/qwen3.6-27b')
    setIsCustomModel(false)
    setCustomModelInput('')
    setApiKey('')
    setMaxTokens('1024')
    setIsFormOpen(true)
  }

  const handleOpenEdit = (cfg: any) => {
    setEditingId(String(cfg.id))
    setProvider(cfg.provider || 'gemini')
    const preset = PROVIDER_PRESETS[cfg.provider] || PROVIDER_PRESETS.gemini
    const found = preset.models.find((m) => m.id === (cfg.modelName || cfg.model))
    if (found) {
      setModel(found.id)
      setIsCustomModel(false)
      setCustomModelInput('')
    } else {
      setIsCustomModel(true)
      setCustomModelInput(cfg.modelName || cfg.model || '')
    }
    setApiKey('')
    setMaxTokens(String(cfg.maxTokens || 1024))
    setIsFormOpen(true)
  }

  const handleTestConnection = async (configId: string) => {
    try {
      const req = new TestLLMConfigRequest({
        id: configId,
        testPrompt: 'ping',
      })
      const resp = await testMutation.mutateAsync(req)
      if (resp.success) {
        toast.success('Koneksi AI Berhasil!')
      } else {
        toast.error(`Koneksi Gagal: ${resp.errorMessage}`)
      }
    } catch (err: any) {
      toast.error(`Uji koneksi gagal: ${err.message || 'Gagal menghubungi API'}`)
    }
  }

  const handleSaveConfig = async () => {
    const finalModel = isCustomModel ? customModelInput.trim() : model
    if (!finalModel) {
      toast.error('Nama/ID Model wajib diisi!')
      return
    }

    if (!editingId && !apiKey.trim() && provider !== 'ollama') {
      toast.error('API Key wajib diisi untuk konfigurasi baru!')
      return
    }

    try {
      if (editingId) {
        const req = new UpdateLLMConfigRequest({
          id: editingId,
          provider,
          modelName: finalModel,
          apiKey: apiKey.trim(),
          maxTokens: Number(maxTokens) || 1024,
          temperature: 0.7,
          enableSkills: true,
          skillsMode: 'prompt',
        })
        await updateMutation.mutateAsync(req)
        toast.success('Konfigurasi LLM berhasil diperbarui')
      } else {
        const req = new CreateLLMConfigRequest({
          provider,
          modelName: finalModel,
          apiKey: apiKey.trim(),
          maxTokens: Number(maxTokens) || 1024,
          temperature: 0.7,
          enableSkills: true,
          skillsMode: 'prompt',
        })
        await createMutation.mutateAsync(req)
        toast.success('Konfigurasi LLM berhasil ditambahkan')
      }
      setIsFormOpen(false)
      setEditingId(null)
      setApiKey('')
    } catch (err: any) {
      toast.error(`Gagal menyimpan konfigurasi: ${err.message}`)
    }
  }

  const handleDeleteConfirmed = async () => {
    if (!configToDelete) return
    try {
      await deleteMutation.mutateAsync(
        new DeleteLLMConfigRequest({ id: String(configToDelete.id) })
      )
      setConfigToDelete(null)
    } catch (err: any) {
      toast.error(`Gagal menghapus LLM: ${err.message}`)
    }
  }

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className='max-h-[90vh] overflow-y-auto sm:max-w-[650px]'>
          <DialogHeader>
            <div className='flex items-center gap-2'>
              <div className='flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary'>
                <Cpu className='h-5 w-5' />
              </div>
              <div>
                <DialogTitle>Konfigurasi Model AI (LLM)</DialogTitle>
                <DialogDescription>
                  Kelola provider AI, API Key, dan model aktif secara langsung tanpa perlu menyentuh file .env.
                </DialogDescription>
              </div>
            </div>
          </DialogHeader>

          {/* DAFTAR KONFIGURASI YANG SUDAH ADA */}
          {!isFormOpen ? (
            <div className='space-y-4 py-2'>
              <div className='flex items-center justify-between'>
                <span className='text-sm font-semibold'>Model Tersimpan</span>
                <Button size='sm' onClick={handleOpenAdd} className='gap-1.5'>
                  <Plus className='h-4 w-4' /> Tambah Provider
                </Button>
              </div>

              {isLoading ? (
                <div className='flex justify-center py-8'>
                  <Loader2 className='h-6 w-6 animate-spin text-muted-foreground' />
                </div>
              ) : configs && configs.length > 0 ? (
                <div className='space-y-2.5'>
                  {configs.map((c: any) => (
                    <div
                      key={c.id}
                      className={`flex items-center justify-between rounded-lg border p-3.5 transition-colors ${
                        c.isActive ? 'border-primary/50 bg-primary/5 shadow-sm' : 'border-border bg-card'
                      }`}
                    >
                      <div className='space-y-1'>
                        <div className='flex items-center gap-2'>
                          <span className='font-medium text-sm'>{c.modelName || c.model}</span>
                          <Badge variant='outline' className='text-[11px] capitalize'>
                            {c.provider}
                          </Badge>
                          {c.isActive && (
                            <Badge className='bg-emerald-600 text-white text-[10px] hover:bg-emerald-600'>
                              Aktif
                            </Badge>
                          )}
                        </div>
                        <div className='flex items-center gap-3 text-xs text-muted-foreground'>
                          <span>Max: {c.maxTokens || 1024} tokens</span>
                          {c.apiKeyMasked && <span>Key: {c.apiKeyMasked}</span>}
                        </div>
                      </div>

                      <div className='flex items-center gap-1.5'>
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => handleTestConnection(String(c.id))}
                          disabled={testMutation.isPending}
                          className='h-8 text-xs gap-1'
                          title='Uji Koneksi Model'
                        >
                          <Zap className='h-3.5 w-3.5 text-amber-500' /> Test
                        </Button>
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => handleOpenEdit(c)}
                          className='h-8 text-xs gap-1'
                          title='Edit Konfigurasi'
                        >
                          <Pencil className='h-3.5 w-3.5' /> Edit
                        </Button>
                        {!c.isActive && (
                          <Button
                            variant='outline'
                            size='sm'
                            onClick={() =>
                              activateMutation.mutate(
                                new ActivateLLMConfigRequest({ id: String(c.id) })
                              )
                            }
                            disabled={activateMutation.isPending}
                            className='h-8 text-xs'
                          >
                            Aktifkan
                          </Button>
                        )}
                        <Button
                          variant='ghost'
                          size='icon'
                          className='h-8 w-8 text-destructive hover:bg-destructive/10'
                          onClick={() => setConfigToDelete(c)}
                          title='Hapus Konfigurasi'
                        >
                          <Trash2 className='h-4 w-4' />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className='rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground'>
                  Belum ada provider LLM yang dikonfigurasi. Klik tombol tambah di atas.
                </div>
              )}
            </div>
          ) : (
            /* FORM TAMBAH / EDIT KONFIGURASI */
            <div className='space-y-4 py-2'>
              <div className='flex items-center gap-2 border-b pb-2 text-sm font-semibold'>
                <Edit3 className='h-4 w-4 text-primary' />
                <span>{editingId ? 'Edit Konfigurasi LLM' : 'Tambah Provider LLM Baru'}</span>
              </div>

              <div className='grid gap-3'>
                <div>
                  <Label className='text-xs font-medium'>1. Provider AI</Label>
                  <Select value={provider} onValueChange={handleProviderChange}>
                    <SelectTrigger className='mt-1'>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {Object.entries(PROVIDER_PRESETS).map(([k, p]) => (
                        <SelectItem key={k} value={k}>
                          {p.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div>
                  <Label className='text-xs font-medium'>2. Pilihan Model</Label>
                  <Select
                    value={isCustomModel ? '__custom__' : model}
                    onValueChange={handleModelSelectChange}
                  >
                    <SelectTrigger className='mt-1'>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {selectedPreset.models.map((m) => (
                        <SelectItem key={m.id} value={m.id}>
                          {m.name}
                        </SelectItem>
                      ))}
                      <SelectItem value='__custom__'>
                        ✏️ Model Kustom Lainnya (Ketik Manual ID Model)...
                      </SelectItem>
                    </SelectContent>
                  </Select>

                  {isCustomModel && (
                    <div className='mt-2'>
                      <Input
                        type='text'
                        placeholder='Contoh: qwen/qwen3.6-27b, openai/gpt-oss-120b, atau deepseek/deepseek-r1'
                        value={customModelInput}
                        onChange={(e) => setCustomModelInput(e.target.value)}
                        className='font-mono text-xs'
                      />
                      <p className='mt-1 text-[11px] text-muted-foreground'>
                        Ketik ID model persis seperti yang disediakan oleh endpoint provider Anda.
                      </p>
                    </div>
                  )}

                  {!isCustomModel && selectedModel && (
                    <p className='mt-1 text-[11px] text-muted-foreground'>{selectedModel.note}</p>
                  )}
                </div>

                <div>
                  <Label className='text-xs font-medium'>
                    3. API Key {editingId && '(Kosongkan jika tidak ingin mengubah)'}
                  </Label>
                  <Input
                    type='password'
                    placeholder={
                      editingId
                        ? '•••••••••••••••• (API Key sudah tersimpan)'
                        : provider === 'gemini'
                        ? 'AIzaSyD...'
                        : provider === 'groq'
                        ? 'gsk_...'
                        : provider === 'openai'
                        ? 'sk-proj-...'
                        : 'Masukkan API Key Anda'
                    }
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value)}
                    className='mt-1 font-mono text-sm'
                  />
                  <p className='mt-1 text-[11px] text-muted-foreground'>
                    API Key akan dienkripsi dengan standar AES-256 dan aman di database.
                  </p>
                </div>

                {/* HARGA RESMI & ESTIMASI BIAYA */}
                <div className='rounded-lg bg-muted/50 p-3 text-xs'>
                  <div className='flex items-center gap-1.5 font-medium text-foreground'>
                    <DollarSign className='h-4 w-4 text-emerald-500' />
                    Estimasi Biaya Token ({selectedModel.name})
                  </div>
                  <div className='mt-2 grid grid-cols-2 gap-2 text-muted-foreground'>
                    <div>
                      Input: <strong className='text-foreground'>${selectedModel.inputPrice}</strong> / 1M token
                    </div>
                    <div>
                      Output: <strong className='text-foreground'>${selectedModel.outputPrice}</strong> / 1M token
                    </div>
                  </div>
                  <div className='mt-1.5 text-[11px] text-emerald-600 dark:text-emerald-400 font-medium'>
                    Simulasi 10.000 chat WhatsApp: ~Rp{' '}
                    {(
                      (12 * selectedModel.inputPrice + 1.8 * selectedModel.outputPrice) *
                      16000
                    ).toLocaleString('id-ID', { maximumFractionDigits: 0 })}{' '}
                    / bulan
                  </div>
                </div>

                <div className='grid grid-cols-2 gap-3'>
                  <div>
                    <Label className='text-xs font-medium'>Max Output Tokens</Label>
                    <Input
                      type='number'
                      value={maxTokens}
                      onChange={(e) => setMaxTokens(e.target.value)}
                      className='mt-1'
                    />
                  </div>
                </div>
              </div>

              <Separator className='my-2' />

              <div className='flex items-center justify-end gap-2'>
                <Button
                  type='button'
                  variant='ghost'
                  size='sm'
                  onClick={() => {
                    setIsFormOpen(false)
                    setEditingId(null)
                  }}
                >
                  Batal
                </Button>
                <Button
                  type='button'
                  size='sm'
                  onClick={handleSaveConfig}
                  disabled={createMutation.isPending || updateMutation.isPending}
                >
                  {(createMutation.isPending || updateMutation.isPending) && (
                    <Loader2 className='mr-1.5 h-4 w-4 animate-spin' />
                  )}
                  {editingId ? 'Simpan Perubahan' : 'Simpan Konfigurasi'}
                </Button>
              </div>
            </div>
          )}

          <DialogFooter className='border-t pt-3'>
            <Button variant='outline' onClick={() => onOpenChange(false)}>
              Tutup
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* MODAL KONFIRMASI HAPUS LLM CONFIG */}
      <ConfirmDialog
        open={!!configToDelete}
        onOpenChange={(open) => {
          if (!open) setConfigToDelete(null)
        }}
        title='Hapus Konfigurasi LLM'
        desc={
          <span>
            Apakah Anda yakin ingin menghapus konfigurasi provider{' '}
            <strong className='font-semibold text-foreground'>{configToDelete?.provider}</strong> (
            {configToDelete?.modelName || configToDelete?.model})? Tindakan ini tidak dapat
            dibatalkan.
          </span>
        }
        confirmText='Hapus Konfigurasi'
        cancelBtnText='Batal'
        destructive
        isLoading={deleteMutation.isPending}
        handleConfirm={handleDeleteConfirmed}
      />
    </>
  )
}
