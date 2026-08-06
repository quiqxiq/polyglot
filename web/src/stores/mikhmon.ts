import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { HotspotProfile, VoucherBatchRequest, VoucherData, VoucherReport } from '../types'
import {
  listHotspotProfilesApi,
  generateVouchersApi,
  renderVoucherHTMLApi,
  getVoucherReportsApi,
  sendVoucherDocumentWAApi,
} from '../api/client'

export const useMikhmonStore = defineStore('mikhmon', () => {
  const selectedDeviceId = ref<string>('')
  const profiles = ref<HotspotProfile[]>([])
  const generatedVouchers = ref<VoucherData[]>([])
  const renderedHTML = ref<string>('')
  const reports = ref<VoucherReport[]>([])
  const loading = ref<boolean>(false)
  const error = ref<string | null>(null)

  async function fetchProfiles() {
    if (!selectedDeviceId.value) {
      profiles.value = []
      return
    }
    try {
      loading.value = true
      error.value = null
      const res = await listHotspotProfilesApi(selectedDeviceId.value)
      profiles.value = res.data || []
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch hotspot profiles'
    } finally {
      loading.value = false
    }
  }

  async function generateVouchers(req: VoucherBatchRequest) {
    if (!selectedDeviceId.value) {
      throw new Error('Pilih router terlebih dahulu')
    }
    try {
      loading.value = true
      error.value = null
      const res = await generateVouchersApi(selectedDeviceId.value, req)
      generatedVouchers.value = res.vouchers || []
      return res.vouchers
    } catch (e: any) {
      error.value = e.message || 'Failed to generate vouchers'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function renderHTML(vouchers?: VoucherData[], templateName: string = 'default') {
    if (!selectedDeviceId.value) {
      throw new Error('Pilih router terlebih dahulu')
    }
    try {
      loading.value = true
      error.value = null
      const listToRender = vouchers || generatedVouchers.value
      const res = await renderVoucherHTMLApi(selectedDeviceId.value, listToRender, templateName)
      renderedHTML.value = res.html || ''
      return res.html
    } catch (e: any) {
      error.value = e.message || 'Failed to render voucher HTML'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchReports(date?: string, month?: string, year?: string) {
    if (!selectedDeviceId.value) {
      reports.value = []
      return
    }
    try {
      loading.value = true
      error.value = null
      const res = await getVoucherReportsApi(selectedDeviceId.value, date, month, year)
      reports.value = res.reports || []
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch sales report'
    } finally {
      loading.value = false
    }
  }

  async function sendVouchersToWA(sessionId: number, recipient: string, fileName: string, htmlContent: string, caption?: string) {
    const fileBase64 = btoa(unescape(encodeURIComponent(htmlContent)))
    return sendVoucherDocumentWAApi(sessionId, recipient, fileName, fileBase64, caption)
  }

  return {
    selectedDeviceId,
    profiles,
    generatedVouchers,
    renderedHTML,
    reports,
    loading,
    error,
    fetchProfiles,
    generateVouchers,
    renderHTML,
    fetchReports,
    sendVouchersToWA,
  }
})
