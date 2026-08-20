import { Fragment, useMemo, useState } from 'react'
import { format, isSameDay } from 'date-fns'
import { toast } from 'sonner'
import {
  TakeOverConversationRequest,
  ResetConversationBotRequest,
  CloseConversationRequest,
  ResetRateLimitRequest,
} from '@/gen/v1/bot_pb'
import {
  MarkChatReadRequest,
  SendWATextMessageRequest,
  ToggleChatBotRequest,
  type WAChat,
  type WAChatMessage,
} from '@/gen/v1/whatsapp_pb'
import {
  ArrowLeft,
  Bot,
  BotOff,
  Check,
  CheckCheck,
  Clock,
  Edit,
  Send,
  Search as SearchIcon,
  MessagesSquare,
  RotateCcw,
  ShieldAlert,
  Smartphone,
  UserCheck,
} from 'lucide-react'
import { cn, getDisplayNameInitials } from '@/lib/utils'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { SSEIndicator } from '@/components/sse-indicator'
import { ThemeSwitch } from '@/components/theme-switch'
import {
  typingKey,
  useWARealtimeStream,
  type ChatPresence,
} from '../whatsapp/api/use-whatsapp-sse'
import {
  useListChatsQuery,
  useGetChatMessagesQuery,
  useMarkChatReadMutation,
  useSendWATextMessageMutation,
  useToggleChatBotMutation,
  useWASessionsQuery,
  useConversationsQuery,
  useConversationContextQuery,
  useTakeOverConversationMutation,
  useResetConversationBotMutation,
  useCloseConversationMutation,
  useRateLimitStatusQuery,
  useResetRateLimitMutation,
} from './api/use-chats'
import { NewChat } from './components/new-chat'

// jid → nomor tampilan ("62812xxxxxxx@s.whatsapp.net" → "62812xxxxxxx")
function formatJid(jid: string): string {
  return jid.split('@')[0] || jid
}

function mediaLabel(mediaType: string, content: string): string {
  if (content) return content
  switch (mediaType) {
    case 'image':
      return '📷 Foto'
    case 'video':
      return '🎬 Video'
    case 'audio':
      return '🎵 Audio'
    case 'document':
      return '📄 Dokumen'
    case 'sticker':
      return '🖼️ Stiker'
    case 'location':
      return '📍 Lokasi'
    case 'contact':
      return '👤 Kontak'
    case 'call':
      return '📞 Panggilan'
    default:
      return '[media]'
  }
}

function chatTime(ts: string): string {
  if (!ts) return ''
  const date = new Date(ts)
  if (Number.isNaN(date.getTime())) return ''
  return isSameDay(date, new Date())
    ? format(date, 'HH:mm')
    : format(date, 'd MMM')
}

// Centang status pengiriman ala WhatsApp pada pesan keluar:
//   - status kosong/sent → ✓ (terkirim)
//   - delivered          → ✓✓ (sampai di device penerima)
//   - read               → ✓✓ biru (dibaca)
// Pesan masuk tidak menampilkan centang.
function MessageStatus({ msg }: { msg: WAChatMessage }) {
  if (!msg.isFromMe) return null
  if (msg.status === 'read') {
    return (
      <CheckCheck size={15} className='text-blue-500' aria-label='Dibaca' />
    )
  }
  if (msg.status === 'delivered') {
    return <CheckCheck size={15} aria-label='Terkirim (✓✓)' />
  }
  if (msg.status === 'pending') {
    return <Clock size={13} aria-label='Mengirim...' />
  }
  return <Check size={15} aria-label='Terkirim (✓)' />
}

function statusLabel(status: string): string {
  switch (status) {
    case 'escalation':
      return 'Diteruskan ke agen'
    case 'done':
      return 'Selesai'
    case 'bot':
      return 'Bot aktif'
    default:
      return status
  }
}

const statusStyles: Record<string, string> = {
  bot: 'bg-emerald-500/15 text-emerald-600',
  escalation: 'bg-amber-500/15 text-amber-600',
  done: 'bg-muted text-muted-foreground',
}

// Tab filter daftar chat: Semua / Belum dibaca / Dibaca / Grup / Baru.
// "Baru" = chat yang belum pernah ditangani bot/agen (belum ada percakapan
// untuk nomor itu) — cocok untuk inbox yang butuh perhatian.
type ChatFilter = 'all' | 'unread' | 'read' | 'group' | 'new'

const CHAT_FILTERS: { key: ChatFilter; label: string }[] = [
  { key: 'all', label: 'Semua' },
  { key: 'unread', label: 'Belum dibaca' },
  { key: 'read', label: 'Dibaca' },
  { key: 'group', label: 'Grup' },
  { key: 'new', label: 'Baru' },
]

// Label indikator typing/recording ala WhatsApp. Untuk grup, nama pengirim
// ditampilkan bila diketahui (fallback: nomor pengirim dari payload SSE).
function typingLabel(p: ChatPresence, senderName?: string): string {
  const who = p.isGroup ? (senderName ? `${senderName} ` : 'Seseorang ') : ''
  if (p.media === 'audio') return `${who}merekam audio…`
  return p.isGroup ? `${who}sedang mengetik…` : 'mengetik…'
}

export function Chats() {
  // Status device & pesan baru ter-update live via SSE (chat_update) —
  // Inbox refresh instan tanpa polling. Status koneksi untuk indikator header.
  // `typing` berisi indikator mengetik/merekam per chat (event chat_presence).
  const { status: sseStatus, typing } = useWARealtimeStream()

  const [search, setSearch] = useState('')
  const [selectedSessionId, setSelectedSessionId] = useState('')
  const [selectedChat, setSelectedChat] = useState<WAChat | null>(null)
  const [mobileSelected, setMobileSelected] = useState(false)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [draft, setDraft] = useState('')
  const [filter, setFilter] = useState<ChatFilter>('all')

  const sessionsQuery = useWASessionsQuery()
  const sessions = useMemo(() => sessionsQuery.data ?? [], [sessionsQuery.data])

  // Fallback ke session pertama bila belum ada pilihan eksplisit.
  const activeSessionId = selectedSessionId || sessions[0]?.id || ''

  const chatsQuery = useListChatsQuery(activeSessionId, search.trim())
  const chats = useMemo(() => chatsQuery.data ?? [], [chatsQuery.data])

  const messagesQuery = useGetChatMessagesQuery(
    activeSessionId,
    selectedChat?.chatJid ?? '',
    Boolean(selectedChat)
  )
  const messages = useMemo(() => messagesQuery.data ?? [], [messagesQuery.data])

  // Temukan percakapan bisnis (bot) milik chat yang dipilih — dicocokkan lewat
  // nomor WA. Percakapan hanya ada setelah bot/agen pernah terlibat.
  const conversationsQuery = useConversationsQuery(activeSessionId)
  const conversations = useMemo(
    () => conversationsQuery.data ?? [],
    [conversationsQuery.data]
  )

  // Nomor yang sudah pernah ditangani bot/agen — dipakai tab filter "Baru".
  const handledPhones = useMemo(
    () => new Set(conversations.map((c) => c.clientPhone)),
    [conversations]
  )

  // Filter diterapkan di sisi client di atas daftar yang sudah dimuat (search
  // tetap server-side; limit dinaikkan agar filter punya cakupan lebih luas).
  const filteredChats = useMemo(() => {
    switch (filter) {
      case 'unread':
        return chats.filter((c) => c.unreadCount > 0)
      case 'read':
        return chats.filter(
          (c) => c.unreadCount === 0 && Boolean(c.lastMessageTime)
        )
      case 'group':
        return chats.filter((c) => c.isGroup)
      case 'new':
        return chats.filter((c) => !handledPhones.has(formatJid(c.chatJid)))
      default:
        return chats
    }
  }, [chats, filter, handledPhones])

  // Jumlah per tab — ditampilkan sebagai badge kecil di chip filter.
  const tabCounts = useMemo(
    () => ({
      all: chats.length,
      unread: chats.filter((c) => c.unreadCount > 0).length,
      read: chats.filter(
        (c) => c.unreadCount === 0 && Boolean(c.lastMessageTime)
      ).length,
      group: chats.filter((c) => c.isGroup).length,
      new: chats.filter((c) => !handledPhones.has(formatJid(c.chatJid))).length,
    }),
    [chats, handledPhones]
  )

  // Indikator mengetik untuk chat yang sedang dibuka (header percakapan).
  const activePres = selectedChat
    ? typing[typingKey(activeSessionId, selectedChat.chatJid)]
    : undefined
  const selectedConv = useMemo(() => {
    if (!selectedChat) return undefined
    const phone = formatJid(selectedChat.chatJid)
    const cleanDigits = phone.replace(/\D/g, '')
    return conversations.find((c) => {
      if (c.clientPhone === phone || c.clientPhone === selectedChat.chatJid) return true
      const cDigits = c.clientPhone.replace(/\D/g, '')
      if (cDigits && cleanDigits) {
        return cDigits === cleanDigits || cDigits.endsWith(cleanDigits) || cleanDigits.endsWith(cDigits)
      }
      return false
    })
  }, [selectedChat, conversations])
  const contextQuery = useConversationContextQuery(
    selectedConv?.id ?? '',
    Boolean(selectedConv)
  )
  const convContext = contextQuery.data

  const selectedPhone = useMemo(() => {
    if (!selectedChat || selectedChat.isGroup) return ''
    return formatJid(selectedChat.chatJid)
  }, [selectedChat])

  const rateLimitQuery = useRateLimitStatusQuery(
    selectedPhone,
    Boolean(selectedPhone)
  )
  const rateLimitStatus = rateLimitQuery.data
  const resetRateLimitMutation = useResetRateLimitMutation()

  const markReadMutation = useMarkChatReadMutation()
  const sendMutation = useSendWATextMessageMutation()
  const toggleChatBotMutation = useToggleChatBotMutation()
  const takeOverMutation = useTakeOverConversationMutation()
  const resetBotMutation = useResetConversationBotMutation()
  const closeConvMutation = useCloseConversationMutation()

  const handleResetRateLimit = async () => {
    if (!selectedPhone) return
    try {
      await resetRateLimitMutation.mutateAsync(
        new ResetRateLimitRequest({ phoneNumber: selectedPhone })
      )
      toast.success(
        `Rate limit dan kuota harian nomor ${selectedPhone} berhasil direset`
      )
    } catch (err: unknown) {
      const errorMsg =
        err instanceof Error ? err.message : 'Terjadi kesalahan'
      toast.error(`Gagal reset rate limit: ${errorMsg}`)
    }
  }

  const handleSelectChat = (chat: WAChat) => {
    setSelectedChat(chat)
    setMobileSelected(true)
    if (chat.unreadCount > 0) {
      markReadMutation.mutate(
        new MarkChatReadRequest({
          sessionId: activeSessionId,
          chatJid: chat.chatJid,
        })
      )
    }
  }

  const handleToggleChatBot = () => {
    if (!selectedChat) return
    const next = !selectedChat.botEnabled
    const prev = selectedChat
    // Optimistic — UI langsung berubah, di-revert bila request gagal.
    const nextChat = prev.clone()
    nextChat.botEnabled = next
    setSelectedChat(nextChat)
    toggleChatBotMutation.mutate(
      new ToggleChatBotRequest({
        sessionId: activeSessionId,
        chatJid: selectedChat.chatJid,
        isActive: next,
      }),
      {
        onError: () => setSelectedChat(prev),
      }
    )
  }

  const handleTakeOver = () => {
    if (selectedChat?.botEnabled) {
      handleToggleChatBot()
    }
    const convId = selectedConv?.id != null ? String(selectedConv.id) : (convContext?.conversationId ? String(convContext.conversationId) : undefined)
    if (convId) {
      takeOverMutation.mutate(
        new TakeOverConversationRequest({ id: convId })
      )
    }
    toast.success('Percakapan dialihkan ke agen CS manual (Bot dinonaktifkan)')
  }

  const handleResetBot = () => {
    if (selectedChat && !selectedChat.botEnabled) {
      handleToggleChatBot()
    }
    const convId = selectedConv?.id != null ? String(selectedConv.id) : (convContext?.conversationId ? String(convContext.conversationId) : undefined)
    if (convId) {
      resetBotMutation.mutate(
        new ResetConversationBotRequest({ id: convId })
      )
    }
    toast.success('Bot AI diaktifkan kembali untuk percakapan ini')
  }

  const handleCloseConversation = () => {
    if (!selectedConv) return
    closeConvMutation.mutate(
      new CloseConversationRequest({ id: selectedConv.id })
    )
  }

  const handleSend = (to: string, content: string) => {
    if (!activeSessionId || !to || (!content.trim() && !selectedChat)) return
    sendMutation.mutate(
      new SendWATextMessageRequest({
        sessionId: activeSessionId,
        recipientPhone: to,
        messageText: content.trim(),
      }),
      {
        onSuccess: () => setDraft(''),
      }
    )
  }  // Pesan dari API ascending (terlama→terbaru). Container chat memakai
  // flex-col-reverse (item pertama dirender di bawah), jadi array dibalik
  // supaya pesan terbaru berada di BAWAH dan scroll menempel di bawah —
  // tanpa pembalikan ini urutan tampil terbalik (terbaru di atas).
  const orderedMessages = useMemo(() => [...messages].reverse(), [messages])

  const groupedMessages = useMemo(() => {
    const groups = new Map<string, WAChatMessage[]>()
    for (const msg of orderedMessages) {
      const key = msg.timestamp ? format(new Date(msg.timestamp), 'd MMM yyyy') : '—'
      const list = groups.get(key) ?? []
      list.push(msg)
      groups.set(key, list)
    }
    return [...groups.entries()]
  }, [orderedMessages])

  return (
    <>
      <Header>
        <Search className='me-auto' />
        <SSEIndicator status={sseStatus} />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main fixed>
        <section className='flex h-full gap-6'>
          {/* ── Kiri: daftar chat ─────────────────────────────── */}
          <div className='flex w-full flex-col gap-2 sm:w-56 lg:w-72 2xl:w-80'>
            <div className='sticky top-0 z-10 -mx-4 bg-background px-4 pb-3 shadow-md sm:static sm:z-auto sm:mx-0 sm:p-0 sm:shadow-none'>
              <div className='flex items-center justify-between py-2'>
                <div className='flex items-center gap-2'>
                  <h1 className='text-2xl font-bold'>Inbox</h1>
                  <MessagesSquare size={20} />
                </div>
                <Button
                  size='icon'
                  variant='ghost'
                  onClick={() => setCreateDialogOpen(true)}
                  className='rounded-lg'
                  title='New message'
                >
                  <Edit size={24} className='stroke-muted-foreground' />
                </Button>
              </div>

              {sessions.length > 1 && (
                <Select
                  value={selectedSessionId}
                  onValueChange={(v) => setSelectedSessionId(v)}
                >
                  <SelectTrigger className='mb-2 h-9 w-full'>
                    <SelectValue placeholder='Pilih perangkat WhatsApp' />
                  </SelectTrigger>
                  <SelectContent>
                    {sessions.map((s) => (
                      <SelectItem key={s.id} value={s.id}>
                        {s.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}

              <label
                className={cn(
                  'focus-within:ring-1 focus-within:ring-ring focus-within:outline-hidden',
                  'flex h-10 w-full items-center space-x-0 rounded-md border border-border ps-2'
                )}
              >
                <SearchIcon size={15} className='me-2 stroke-slate-500' />
                <span className='sr-only'>Search</span>
                <input
                  type='text'
                  className='w-full flex-1 bg-inherit text-sm focus-visible:outline-hidden'
                  placeholder='Search chat...'
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
              </label>

              <div className='mt-2 flex flex-wrap gap-1.5'>
                {CHAT_FILTERS.map(({ key, label }) => {
                  const count = tabCounts[key]
                  return (
                    <button
                      key={key}
                      type='button'
                      onClick={() => setFilter(key)}
                      className={cn(
                        'inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-medium transition-colors',
                        filter === key
                          ? 'bg-primary text-primary-foreground'
                          : 'bg-muted text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                      )}
                    >
                      {label}
                      {count > 0 && (
                        <span
                          className={cn(
                            'text-[10px]',
                            filter === key
                              ? 'text-primary-foreground/80'
                              : 'text-muted-foreground/70'
                          )}
                        >
                          {count}
                        </span>
                      )}
                    </button>
                  )
                })}
              </div>
            </div>

            <ScrollArea className='-mx-3 h-full overflow-scroll p-3'>
              {!sessions.length && (
                <p className='px-3 py-6 text-center text-sm text-muted-foreground'>
                  Belum ada perangkat WhatsApp terhubung.
                </p>
              )}
              {sessions.length > 0 && !filteredChats.length && (
                <p className='px-3 py-6 text-center text-sm text-muted-foreground'>
                  {search
                    ? 'Tidak ada chat yang cocok.'
                    : filter === 'all'
                      ? 'Belum ada percakapan. Kirim pesan untuk memulai.'
                      : 'Tidak ada chat pada filter ini.'}
                </p>
              )}
              {filteredChats.map((chat) => {
                const displayName = chat.displayName || formatJid(chat.chatJid)
                const preview = chat.lastMessagePreview || '[media]'
                // Indikator mengetik/merekam untuk chat ini (dari SSE).
                const pres = typing[typingKey(activeSessionId, chat.chatJid)]
                const presName =
                  pres?.isGroup && pres.senderJid
                    ? formatJid(pres.senderJid)
                    : undefined
                return (
                  <Fragment key={chat.id}>
                    <button
                      type='button'
                      className={cn(
                        'group hover:bg-accent hover:text-accent-foreground',
                        'flex w-full items-center gap-2 rounded-md px-2 py-2 text-start text-sm',
                        selectedChat?.chatJid === chat.chatJid && 'sm:bg-muted'
                      )}
                      onClick={() => handleSelectChat(chat)}
                    >
                      <Avatar className='size-10 shrink-0'>
                        {chat.isGroup ? (
                          <AvatarFallback className='bg-primary/15 text-primary'>
                            {getDisplayNameInitials(displayName)}
                          </AvatarFallback>
                        ) : (
                          <AvatarFallback>
                            {getDisplayNameInitials(displayName)}
                          </AvatarFallback>
                        )}
                      </Avatar>
                      <div className='min-w-0 flex-1'>
                        <div className='flex items-baseline justify-between gap-2'>
                          <span className='truncate font-medium'>
                            {displayName}
                          </span>
                          <span className='shrink-0 text-xs text-muted-foreground'>
                            {chatTime(chat.lastMessageTime)}
                          </span>
                        </div>
                        <div className='flex items-center justify-between gap-2'>
                          <span
                            className={cn(
                              'line-clamp-1 truncate text-ellipsis',
                              pres
                                ? 'font-medium text-emerald-600'
                                : 'text-muted-foreground'
                            )}
                          >
                            {pres ? typingLabel(pres, presName) : preview}
                          </span>
                          <span className='flex shrink-0 items-center gap-1'>
                            {!chat.botEnabled && (
                              <span
                                className='inline-flex items-center rounded bg-muted px-1 py-0.5 text-[10px] font-medium text-muted-foreground'
                                title='Bot nonaktif untuk chat ini'
                              >
                                <BotOff size={11} className='me-0.5' /> off
                              </span>
                            )}
                            {chat.unreadCount > 0 && (
                              <span className='inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1.5 text-xs font-medium text-primary-foreground'>
                                {chat.unreadCount}
                              </span>
                            )}
                          </span>
                        </div>
                      </div>
                    </button>
                    <Separator className='my-1' />
                  </Fragment>
                )
              })}
            </ScrollArea>
          </div>

          {/* ── Kanan: percakapan ─────────────────────────────── */}
          {selectedChat ? (
            <div
              className={cn(
                'absolute inset-0 start-full z-50 hidden w-full flex-1 flex-col border bg-background shadow-xs sm:static sm:z-auto sm:flex sm:rounded-md',
                mobileSelected && 'inset-s-0 flex'
              )}
            >
              <div className='mb-1 flex flex-none justify-between bg-card p-4 shadow-lg sm:rounded-t-md'>
                <div className='flex gap-3'>
                  <Button
                    size='icon'
                    variant='ghost'
                    className='-ms-2 h-full sm:hidden'
                    onClick={() => setMobileSelected(false)}
                  >
                    <ArrowLeft className='rtl:rotate-180' />
                  </Button>
                  <div className='flex items-center gap-2 lg:gap-4'>
                    <Avatar className='size-9 lg:size-11'>
                      <AvatarFallback>
                        {getDisplayNameInitials(
                          selectedChat.displayName ||
                            formatJid(selectedChat.chatJid)
                        )}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <span className='col-start-2 row-span-2 text-sm font-medium lg:text-base'>
                        {selectedChat.displayName ||
                          formatJid(selectedChat.chatJid)}
                      </span>
                      <span className='col-start-2 row-span-2 row-start-2 block max-w-32 text-xs text-nowrap text-ellipsis lg:max-w-none lg:text-sm'>
                        {activePres ? (
                          <span className='font-medium text-emerald-600'>
                            {typingLabel(
                              activePres,
                              activePres.isGroup && activePres.senderJid
                                ? formatJid(activePres.senderJid)
                                : undefined
                            )}
                          </span>
                        ) : (
                          <span className='text-muted-foreground'>
                            {selectedChat.isGroup
                              ? 'Grup'
                              : formatJid(selectedChat.chatJid)}
                          </span>
                        )}
                      </span>
                    </div>
                  </div>
                </div>
                <div className='-me-1 flex items-center gap-1 lg:gap-2'>
                  <Button
                    size='icon'
                    variant='ghost'
                    className='h-9 w-9 rounded-md'
                    title={
                      selectedChat.botEnabled
                        ? 'Nonaktifkan bot untuk chat ini'
                        : 'Aktifkan bot untuk chat ini'
                    }
                    disabled={toggleChatBotMutation.isPending}
                    onClick={handleToggleChatBot}
                  >
                    {selectedChat.botEnabled ? (
                      <Bot size={18} className='text-emerald-500' />
                    ) : (
                      <BotOff size={18} className='text-muted-foreground' />
                    )}
                  </Button>
                </div>
              </div>

              {(convContext || selectedPhone) && (
                <div className='flex flex-none flex-wrap items-center gap-x-3 gap-y-1 border-b bg-card px-4 py-2 text-xs'>
                  {convContext && (
                    <span
                      className={cn(
                        'rounded-full px-2 py-0.5 font-medium',
                        statusStyles[convContext.status] ??
                          'bg-muted text-muted-foreground'
                      )}
                    >
                      {statusLabel(convContext.status)}
                    </span>
                  )}
                  {rateLimitStatus?.isMuted && (
                    <span className='inline-flex items-center gap-1 rounded-full bg-destructive/15 px-2 py-0.5 font-medium text-destructive'>
                      <ShieldAlert size={12} /> Rate Limited (Muted)
                    </span>
                  )}
                  {rateLimitStatus &&
                    rateLimitStatus.dailyChatCount >=
                      rateLimitStatus.dailyQuotaLimit &&
                    rateLimitStatus.dailyQuotaLimit > 0 && (
                      <span className='inline-flex items-center gap-1 rounded-full bg-amber-500/15 px-2 py-0.5 font-medium text-amber-700 dark:text-amber-300'>
                        <ShieldAlert size={12} /> Kuota AI Habis (
                        {rateLimitStatus.dailyChatCount}/
                        {rateLimitStatus.dailyQuotaLimit})
                      </span>
                    )}
                  {rateLimitStatus?.isWhitelisted && (
                    <span className='inline-flex items-center gap-1 rounded-full bg-blue-500/15 px-2 py-0.5 font-medium text-blue-700 dark:text-blue-300'>
                      ⭐ Whitelist
                    </span>
                  )}
                  {rateLimitStatus &&
                    !rateLimitStatus.isMuted &&
                    !rateLimitStatus.isWhitelisted && (
                      <span className='text-muted-foreground'>
                        Kuota AI:{' '}
                        <span className='font-medium text-foreground'>
                          {rateLimitStatus.dailyChatCount}
                        </span>
                        /{rateLimitStatus.dailyQuotaLimit}
                      </span>
                    )}
                  {convContext && (
                    <>
                      <span className='text-muted-foreground'>
                        Tokens:{' '}
                        <span className='font-medium text-foreground'>
                          {convContext.totalTokenIn.toLocaleString('id-ID')}
                        </span>{' '}
                        in /{' '}
                        <span className='font-medium text-foreground'>
                          {convContext.totalTokenOut.toLocaleString('id-ID')}
                        </span>{' '}
                        out
                      </span>
                      <span className='text-muted-foreground'>
                        {convContext.totalLlmCalls.toLocaleString('id-ID')}{' '}
                        balasan LLM
                      </span>
                    </>
                  )}
                  {convContext?.summary && (
                    <p
                      className='w-full truncate text-muted-foreground'
                      title={convContext.summary}
                    >
                      <span className='font-medium text-foreground not-italic'>
                        Ringkasan bot:
                      </span>{' '}
                      {convContext.summary}
                    </p>
                  )}
                  <div className='ms-auto flex items-center gap-2'>
                    {/* Tombol Reset Rate Limit & Kuota Harian (Selalu Tersedia) */}
                    {selectedPhone && (
                      <Button
                        size='sm'
                        variant={rateLimitStatus?.isMuted ? 'destructive' : 'outline'}
                        className={cn(
                          'gap-1.5 text-xs',
                          rateLimitStatus?.isMuted
                            ? 'animate-pulse'
                            : 'border-amber-300 text-amber-700 hover:bg-amber-50 dark:border-amber-700 dark:text-amber-300 dark:hover:bg-amber-950/50'
                        )}
                        disabled={resetRateLimitMutation.isPending}
                        onClick={handleResetRateLimit}
                        title='Reset pembatasan spam (mute 1 jam/24 jam) dan kuota percakapan AI nomor ini'
                      >
                        <RotateCcw size={13} /> Reset Limit
                      </Button>
                    )}

                    {/* Tombol Ambil Alih CS vs Aktifkan AI */}
                    {selectedChat?.botEnabled && convContext?.status !== 'escalation' ? (
                      <Button
                        size='sm'
                        variant='outline'
                        className='gap-1.5 text-xs'
                        disabled={takeOverMutation.isPending || toggleChatBotMutation.isPending}
                        onClick={handleTakeOver}
                        title='Hentikan bot AI dan alihkan percakapan ke CS manual'
                      >
                        <UserCheck size={14} className='text-amber-600' /> Ambil Alih CS
                      </Button>
                    ) : (
                      <Button
                        size='sm'
                        variant='outline'
                        className='gap-1.5 text-xs border-emerald-300 text-emerald-700 hover:bg-emerald-50 dark:border-emerald-700 dark:text-emerald-300 dark:hover:bg-emerald-950/50'
                        disabled={resetBotMutation.isPending || toggleChatBotMutation.isPending}
                        onClick={handleResetBot}
                        title='Kembalikan percakapan ke mode bot AI otomatis'
                      >
                        <Bot size={14} className='text-emerald-600' /> Aktifkan AI
                      </Button>
                    )}

                    {convContext && convContext.status !== 'done' && convContext.status !== 'closed' && (
                      <Button
                        size='sm'
                        variant='ghost'
                        className='text-xs text-muted-foreground hover:text-foreground'
                        disabled={closeConvMutation.isPending}
                        onClick={handleCloseConversation}
                        title='Tandai percakapan selesai'
                      >
                        Tutup Chat
                      </Button>
                    )}
                  </div>
                </div>
              )}

              <div className='flex flex-1 flex-col gap-2 rounded-md px-4 pt-0 pb-4'>
                <div className='flex size-full flex-1'>
                  <div className='chat-text-container relative -me-4 flex flex-1 flex-col overflow-y-hidden'>
                    <div className='chat-flex flex h-40 w-full grow flex-col-reverse justify-start gap-4 overflow-y-auto py-2 pe-4 pb-4'>
                      {groupedMessages.length === 0 && (
                        <p className='py-6 text-center text-sm text-muted-foreground'>
                          Belum ada pesan di percakapan ini.
                        </p>
                      )}
                      {groupedMessages.map(([day, msgs]) => (
                        <Fragment key={day}>
                          <div className='text-center text-xs text-muted-foreground'>
                            {day}
                          </div>
                          {msgs.map((msg, index) => (
                            <div
                              key={`${msg.id}-${index}`}
                              className={cn(
                                'chat-box max-w-72 px-3 py-2 wrap-break-word shadow-lg',
                                msg.isFromMe
                                  ? 'self-end rounded-[16px_16px_0_16px] bg-primary/90 text-primary-foreground/75'
                                  : 'self-start rounded-[16px_16px_16px_0] bg-muted'
                              )}
                            >
                              {msg.content
                                ? msg.content
                                : mediaLabel(msg.mediaType ?? '', msg.content)}
                              <span
                                className={cn(
                                  'mt-1 flex items-center gap-1 text-xs font-light text-foreground/75 italic',
                                  msg.isFromMe &&
                                    'justify-end text-primary-foreground/85'
                                )}
                              >
                                {msg.timestamp
                                  ? format(new Date(msg.timestamp), 'HH:mm')
                                  : ''}
                                <MessageStatus msg={msg} />
                              </span>
                            </div>
                          ))}
                        </Fragment>
                      ))}
                    </div>
                  </div>
                </div>

                <form
                  className='flex w-full flex-none gap-2'
                  onSubmit={(e) => {
                    e.preventDefault()
                    handleSend(selectedChat.chatJid, draft)
                  }}
                >
                  <div className='flex flex-1 items-center gap-2 rounded-md border border-input bg-card px-2 py-1 focus-within:ring-1 focus-within:ring-ring focus-within:outline-hidden lg:gap-4'>
                    <label className='flex-1'>
                      <span className='sr-only'>Chat Text Box</span>
                      <input
                        type='text'
                        placeholder='Type your messages...'
                        className='h-8 w-full bg-inherit focus-visible:outline-hidden'
                        value={draft}
                        onChange={(e) => setDraft(e.target.value)}
                      />
                    </label>
                    <Button
                      type='submit'
                      variant='ghost'
                      size='icon'
                      className='hidden sm:inline-flex'
                      disabled={!draft.trim() || sendMutation.isPending}
                    >
                      <Send size={20} />
                    </Button>
                  </div>
                  <Button
                    type='submit'
                    className='h-full sm:hidden'
                    disabled={!draft.trim() || sendMutation.isPending}
                  >
                    <Send size={18} /> Send
                  </Button>
                </form>
              </div>
            </div>
          ) : (
            <div
              className={cn(
                'absolute inset-0 start-full z-50 hidden w-full flex-1 flex-col justify-center rounded-md border bg-card shadow-xs sm:static sm:z-auto sm:flex'
              )}
            >
              <div className='flex flex-col items-center space-y-6'>
                <div className='flex size-16 items-center justify-center rounded-full border-2 border-border'>
                  <Smartphone className='size-8' />
                </div>
                <div className='space-y-2 text-center'>
                  <h1 className='text-xl font-semibold'>Your messages</h1>
                  <p className='text-sm text-muted-foreground'>
                    Pilih percakapan di kiri atau kirim pesan untuk memulai.
                  </p>
                </div>
                <Button onClick={() => setCreateDialogOpen(true)}>
                  Send message
                </Button>
              </div>
            </div>
          )}
        </section>

        <NewChat
          open={createDialogOpen}
          onOpenChange={setCreateDialogOpen}
          pending={sendMutation.isPending}
          onSend={(phone, messageText) => {
            handleSend(phone, messageText)
            setCreateDialogOpen(false)
          }}
        />
      </Main>
    </>
  )
}
