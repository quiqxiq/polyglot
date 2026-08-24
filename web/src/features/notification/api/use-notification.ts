import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  type NotificationTemplate,
  type TestSendRequest,
} from '@/gen/v1/notification_pb'
import { notificationClient } from '@/lib/api-client'
import { notificationKeys } from './keys'

export function useNotificationTemplatesQuery(activeOnly = false) {
  return useQuery({
    queryKey: notificationKeys.templates.list(activeOnly),
    queryFn: async () => {
      const res = await notificationClient.listTemplates({ activeOnly })
      return res.templates
    },
  })
}

export function useNotificationTemplateQuery(templateKey: string) {
  return useQuery({
    queryKey: notificationKeys.templates.detail(templateKey),
    queryFn: async () => {
      const res = await notificationClient.getTemplate({ templateKey })
      return res.template
    },
    enabled: Boolean(templateKey),
  })
}

export function useNotificationsQuery(customerId = '', status = '', limit = 50) {
  return useQuery({
    queryKey: notificationKeys.queue.list(customerId, status, limit),
    queryFn: async () => {
      const res = await notificationClient.listNotifications({
        customerId,
        status,
        limit,
      })
      return res.notifications
    },
  })
}

export function usePendingCountQuery() {
  return useQuery({
    queryKey: notificationKeys.queue.pendingCount(),
    queryFn: async () => {
      return await notificationClient.pendingCount({})
    },
    refetchInterval: 10000,
  })
}

export function useSaveNotificationTemplateMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (template: Partial<NotificationTemplate>) => {
      return await notificationClient.saveTemplate({ template: template as NotificationTemplate })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationKeys.templates.all() })
    },
  })
}

export function useMarkNotificationSentMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      return await notificationClient.markNotificationSent({ id })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationKeys.queue.all() })
    },
  })
}

export function useMarkNotificationFailedMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, errorMessage }: { id: string; errorMessage: string }) => {
      return await notificationClient.markNotificationFailed({ id, errorMessage })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationKeys.queue.all() })
    },
  })
}

export function useTestSendNotificationMutation() {
  return useMutation({
    mutationFn: async (req: TestSendRequest) => {
      return await notificationClient.testSend(req)
    },
  })
}
