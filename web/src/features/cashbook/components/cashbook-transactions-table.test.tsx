import { describe, expect, it } from 'vitest'
import { render } from 'vitest-browser-react'
import { CashAccount, CashCategory, CashTransaction } from '@/gen/v1/cashbook_pb'
import { CashbookProvider } from './cashbook-provider'
import { CashbookTransactionsTable } from './cashbook-transactions-table'

const accounts = [
  new CashAccount({
    id: 'ca-1',
    accountCode: '1001-KAS',
    name: 'Kas Utama Kantor',
    type: 'CASH',
    isActive: true,
  }),
  new CashAccount({
    id: 'ca-2',
    accountCode: '1002-BCA',
    name: 'Bank BCA',
    type: 'BANK',
    isActive: true,
  }),
]

const categories = [
  new CashCategory({
    id: 'cc-1',
    name: 'Pendapatan Tagihan',
    type: 'INCOME',
    isActive: true,
  }),
  new CashCategory({
    id: 'cc-2',
    name: 'Listrik & Utilitas',
    type: 'EXPENSE',
    isActive: true,
  }),
]

const transactions = [
  new CashTransaction({
    id: 'trx-001',
    transactionNo: 'TRX-202609-000001',
    accountId: 'ca-1',
    categoryId: 'cc-1',
    direction: 'IN',
    amount: 250000,
    sourceType: 'PAYMENT',
    description: 'Pembayaran tagihan INV-202609-001 Budi',
  }),
  new CashTransaction({
    id: 'trx-002',
    transactionNo: 'TRX-202609-000002',
    accountId: 'ca-2',
    categoryId: 'cc-2',
    direction: 'OUT',
    amount: 150000,
    sourceType: 'EXPENSE',
    description: 'Beli token listrik kantor POP',
  }),
]

function renderTable(props: {
  data: CashTransaction[]
  accounts: CashAccount[]
  categories: CashCategory[]
  isLoading?: boolean
}) {
  return render(
    <CashbookProvider>
      <CashbookTransactionsTable {...props} />
    </CashbookProvider>
  )
}

describe('CashbookTransactionsTable', () => {
  it('renders transactions with resolved account and category names', async () => {
    const { getByText } = await renderTable({
      data: transactions,
      accounts,
      categories,
    })

    await expect.element(getByText('TRX-202609-000001')).toBeInTheDocument()
    await expect.element(getByText('TRX-202609-000002')).toBeInTheDocument()
    await expect.element(getByText('Kas Utama Kantor')).toBeInTheDocument()
    await expect.element(getByText('Bank BCA')).toBeInTheDocument()
    await expect.element(getByText('Pendapatan Tagihan')).toBeInTheDocument()
    await expect.element(getByText('Listrik & Utilitas')).toBeInTheDocument()
  })

  it('renders direction and source badges', async () => {
    const { getByText } = await renderTable({
      data: transactions,
      accounts,
      categories,
    })

    await expect.element(getByText('Masuk', { exact: true })).toBeInTheDocument()
    await expect.element(getByText('Keluar', { exact: true })).toBeInTheDocument()
    await expect.element(getByText('Tagihan', { exact: true })).toBeInTheDocument()
    await expect.element(getByText('Operasional', { exact: true })).toBeInTheDocument()
  })

  it('shows empty state when there are no transactions', async () => {
    const { getByText } = await renderTable({
      data: [],
      accounts,
      categories,
    })

    await expect.element(getByText('Belum ada mutasi transaksi kas')).toBeInTheDocument()
  })
})
