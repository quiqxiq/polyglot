import { pppClient } from '@/lib/api-client'
import {
  CreatePPPSecretRequest,
  DeletePPPSecretRequest,
  ListPPPSecretsRequest,
  SetPPPSecretDisabledRequest,
  UpdatePPPSecretRequest,
} from '@/gen/v1/ppp_pb'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { pppKeys } from './keys'

export function usePPPSecretsQuery(deviceId?: string, nameFilter?: string) {
  return useQuery({
    queryKey: [...pppKeys.secrets(deviceId), nameFilter],
    queryFn: async () => {
      if (!deviceId) return []
      const res = await pppClient.listSecrets(
        new ListPPPSecretsRequest({
          deviceId,
          nameFilter: nameFilter || '',
        })
      )
      return res.secrets
    },
    enabled: !!deviceId,
  })
}

export type CreatePPPSecretParams = {
  deviceId: string
  name: string
  password?: string
  profile?: string
  service?: string
  localAddress?: string
  remoteAddress?: string
  comment?: string
  disabled?: boolean
  callerId?: string
}

export function useCreatePPPSecretMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: CreatePPPSecretParams) => {
      return await pppClient.createSecret(new CreatePPPSecretRequest(params))
    },
    onSuccess: (_, variables) => {
      toast.success('PPPoE Secret created successfully')
      queryClient.invalidateQueries({
        queryKey: pppKeys.secrets(variables.deviceId),
      })
      queryClient.invalidateQueries({
        queryKey: pppKeys.inactive(variables.deviceId),
      })
    },
    onError: (err: Error) => {
      toast.error(`Failed to create secret: ${err.message}`)
    },
  })
}

export type UpdatePPPSecretParams = {
  deviceId: string
  rosId: string
  name?: string
  password?: string
  profile?: string
  service?: string
  localAddress?: string
  remoteAddress?: string
  comment?: string
  callerId?: string
}

export function useUpdatePPPSecretMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: UpdatePPPSecretParams) => {
      return await pppClient.updateSecret(new UpdatePPPSecretRequest(params))
    },
    onSuccess: (_, variables) => {
      toast.success('PPPoE Secret updated successfully')
      queryClient.invalidateQueries({
        queryKey: pppKeys.secrets(variables.deviceId),
      })
      queryClient.invalidateQueries({
        queryKey: pppKeys.inactive(variables.deviceId),
      })
    },
    onError: (err: Error) => {
      toast.error(`Failed to update secret: ${err.message}`)
    },
  })
}

export type DeletePPPSecretParams = {
  deviceId: string
  rosId: string
}

export function useDeletePPPSecretMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: DeletePPPSecretParams) => {
      return await pppClient.deleteSecret(new DeletePPPSecretRequest(params))
    },
    onSuccess: (_, variables) => {
      toast.success('PPPoE Secret deleted successfully')
      queryClient.invalidateQueries({
        queryKey: pppKeys.secrets(variables.deviceId),
      })
      queryClient.invalidateQueries({
        queryKey: pppKeys.inactive(variables.deviceId),
      })
    },
    onError: (err: Error) => {
      toast.error(`Failed to delete secret: ${err.message}`)
    },
  })
}

export type SetPPPSecretDisabledParams = {
  deviceId: string
  rosId: string
  disabled: boolean
}

export function useSetPPPSecretDisabledMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: SetPPPSecretDisabledParams) => {
      return await pppClient.setSecretDisabled(new SetPPPSecretDisabledRequest(params))
    },
    onSuccess: (_, variables) => {
      toast.success(variables.disabled ? 'Secret disabled' : 'Secret enabled')
      queryClient.invalidateQueries({
        queryKey: pppKeys.secrets(variables.deviceId),
      })
      queryClient.invalidateQueries({
        queryKey: pppKeys.inactive(variables.deviceId),
      })
    },
    onError: (err: Error) => {
      toast.error(`Failed to change secret status: ${err.message}`)
    },
  })
}

export function useBulkDeletePPPSecretsMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      deviceId,
      rosIds,
    }: {
      deviceId: string
      rosIds: string[]
    }) => {
      for (const rosId of rosIds) {
        await pppClient.deleteSecret(
          new DeletePPPSecretRequest({ deviceId, rosId })
        )
      }
      return rosIds.length
    },
    onSuccess: (count, variables) => {
      toast.success(`${count} secrets deleted successfully`)
      queryClient.invalidateQueries({
        queryKey: pppKeys.secrets(variables.deviceId),
      })
      queryClient.invalidateQueries({
        queryKey: pppKeys.inactive(variables.deviceId),
      })
    },
    onError: (err: Error) => {
      toast.error(`Failed to delete secrets: ${err.message}`)
    },
  })
}

export function useBulkSetPPPSecretsDisabledMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      deviceId,
      rosIds,
      disabled,
    }: {
      deviceId: string
      rosIds: string[]
      disabled: boolean
    }) => {
      for (const rosId of rosIds) {
        await pppClient.setSecretDisabled(
          new SetPPPSecretDisabledRequest({ deviceId, rosId, disabled })
        )
      }
      return rosIds.length
    },
    onSuccess: (count, variables) => {
      toast.success(
        `${count} secrets ${variables.disabled ? 'disabled' : 'enabled'} successfully`
      )
      queryClient.invalidateQueries({
        queryKey: pppKeys.secrets(variables.deviceId),
      })
      queryClient.invalidateQueries({
        queryKey: pppKeys.inactive(variables.deviceId),
      })
    },
    onError: (err: Error) => {
      toast.error(`Failed to update secrets status: ${err.message}`)
    },
  })
}
