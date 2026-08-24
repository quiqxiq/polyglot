import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  type SubmitRegistrationRequest,
  type ApproveRegistrationRequest,
  type ScheduleInstallRequest,
  type MarkInstalledRequest,
  type RejectRegistrationRequest,
  type CancelRegistrationRequest,
  type ConvertRegistrationRequest,
} from '@/gen/v1/registration_pb'
import { registrationClient } from '@/lib/api-client'
import { registrationKeys } from './keys'

export function useRegistrationsQuery(status = '', phone = '') {
  return useQuery({
    queryKey: registrationKeys.list(status, phone),
    queryFn: async () => {
      const res = await registrationClient.listRegistrations({
        status,
        phone,
      })
      return res.registrations
    },
  })
}

export function useRegistrationQuery(id: string) {
  return useQuery({
    queryKey: registrationKeys.detail(id),
    queryFn: async () => {
      const res = await registrationClient.getRegistration({ id })
      return res.registration
    },
    enabled: Boolean(id),
  })
}

export function useSubmitRegistrationMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: SubmitRegistrationRequest) => {
      return await registrationClient.submitRegistration(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: registrationKeys.all })
    },
  })
}

export function useApproveRegistrationMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ApproveRegistrationRequest) => {
      return await registrationClient.approveRegistration(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: registrationKeys.all })
      queryClient.invalidateQueries({ queryKey: registrationKeys.detail(vars.id) })
    },
  })
}

export function useScheduleInstallMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ScheduleInstallRequest) => {
      return await registrationClient.scheduleInstall(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: registrationKeys.all })
      queryClient.invalidateQueries({ queryKey: registrationKeys.detail(vars.id) })
    },
  })
}

export function useMarkInstalledMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: MarkInstalledRequest) => {
      return await registrationClient.markInstalled(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: registrationKeys.all })
      queryClient.invalidateQueries({ queryKey: registrationKeys.detail(vars.id) })
    },
  })
}

export function useRejectRegistrationMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: RejectRegistrationRequest) => {
      return await registrationClient.rejectRegistration(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: registrationKeys.all })
      queryClient.invalidateQueries({ queryKey: registrationKeys.detail(vars.id) })
    },
  })
}

export function useCancelRegistrationMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CancelRegistrationRequest) => {
      return await registrationClient.cancelRegistration(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: registrationKeys.all })
      queryClient.invalidateQueries({ queryKey: registrationKeys.detail(vars.id) })
    },
  })
}

export function useConvertRegistrationMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ConvertRegistrationRequest) => {
      return await registrationClient.convertRegistration(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: registrationKeys.all })
      queryClient.invalidateQueries({ queryKey: registrationKeys.detail(vars.id) })
      queryClient.invalidateQueries({ queryKey: ['customers'] })
      queryClient.invalidateQueries({ queryKey: ['billing'] })
    },
  })
}
