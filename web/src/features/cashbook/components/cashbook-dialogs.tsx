import { CashbookTransactionDialog } from './cashbook-transaction-dialog'
import { CashbookMutateAccountDialog } from './cashbook-mutate-account-dialog'
import { CashbookMutateCategoryDialog } from './cashbook-mutate-category-dialog'
import { useCashbook, type CashbookDialogType } from './cashbook-provider'

export function CashbookDialogs() {
  const { open, setOpen, currentAccount, currentCategory } = useCashbook()

  const close = (dialog: CashbookDialogType) => {
    setOpen(dialog)
  }

  return (
    <>
      {/* Dialog Catat Mutasi Kas Manual */}
      {open === 'create-transaction' && (
        <CashbookTransactionDialog
          open={open === 'create-transaction'}
          onOpenChange={() => close('create-transaction')}
        />
      )}

      {/* Dialog Tambah Rekening */}
      {open === 'create-account' && (
        <CashbookMutateAccountDialog
          open={open === 'create-account'}
          onOpenChange={() => close('create-account')}
        />
      )}

      {/* Dialog Edit Rekening */}
      {currentAccount && open === 'edit-account' && (
        <CashbookMutateAccountDialog
          key={`edit-account-${currentAccount.id}`}
          open={open === 'edit-account'}
          onOpenChange={() => close('edit-account')}
          currentAccount={currentAccount}
        />
      )}

      {/* Dialog Tambah Kategori */}
      {open === 'create-category' && (
        <CashbookMutateCategoryDialog
          open={open === 'create-category'}
          onOpenChange={() => close('create-category')}
        />
      )}

      {/* Dialog Edit Kategori */}
      {currentCategory && open === 'edit-category' && (
        <CashbookMutateCategoryDialog
          key={`edit-category-${currentCategory.id}`}
          open={open === 'edit-category'}
          onOpenChange={() => close('edit-category')}
          currentCategory={currentCategory}
        />
      )}
    </>
  )
}
