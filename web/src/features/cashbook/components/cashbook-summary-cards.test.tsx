import { describe, expect, it, vi } from 'vitest'
import { render } from 'vitest-browser-react'
import { CashAccount, CashTransaction } from '@/gen/v1/cashbook_pb'
import { CashbookProvider } from './cashbook-provider'
import { CashbookSummaryCards } from './cashbook-summary-cards'

vi.mock('../api/use-cashbook', () => ({
  useCashAccountsQuery: () => ({
    data: [
      new CashAccount({
        id: 'ca-1',
        name: 'Kas Utama',
        type: 'CASH',
        isActive: true,
      }),
      new CashAccount({
        id: 'ca-2',
        name: 'Bank BCA',
        type: 'BANK',
        isActive: true,
      }),
    ],
  }),
  useCashBalancesQuery: () => ({
    data: {
      'ca-1': 500000,
      'ca-2': 1500000,
    },
  }),
  useCashTransactionsQuery: () => ({
    data: [
      new CashTransaction({
        id: 'trx-1',
        direction: 'IN',
        amount: 1000000,
      }),
      new CashTransaction({
        id: 'trx-2',
        direction: 'OUT',
        amount: 300000,
      }),
    ],
  }),
}))

describe('CashbookSummaryCards', () => {
  it('renders correct total balance and cashflow metrics', async () => {
    const { getByText } = await render(
      <CashbookProvider>
        <CashbookSummaryCards />
      </CashbookProvider>
    )

    await expect.element(getByText('Total Saldo Bersih')).toBeInTheDocument()
    await expect.element(getByText('Pemasukan Kas')).toBeInTheDocument()
    await expect.element(getByText('Pengeluaran Kas')).toBeInTheDocument()
    await expect.element(getByText('Arus Kas Bersih (Net)')).toBeInTheDocument()
  })
})
