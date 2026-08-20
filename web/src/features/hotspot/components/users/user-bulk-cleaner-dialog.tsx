import { useState } from 'react'
import { Trash2, AlertTriangle, UserX, Clock, Tag } from 'lucide-react'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { useDeleteHotspotUsersMutation } from '../../api/use-hotspot-users'
import { useHotspotProfilesQuery } from '../../api/use-hotspot-profiles'
import { useHotspotUsersQuery } from '../../api/use-hotspot-users'
import { useDeviceStore } from '@/stores/device-store'

type UserBulkCleanerDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function UserBulkCleanerDialog({
  open,
  onOpenChange,
}: UserBulkCleanerDialogProps) {
  const { selectedDeviceId } = useDeviceStore()
  const [activeTab, setActiveTab] = useState<'profile' | 'comment' | 'expired'>('profile')
  const [selectedProfile, setSelectedProfile] = useState('')
  const [selectedComment, setSelectedComment] = useState('')
  const [customComment, setCustomComment] = useState('')

  const { data: profiles = [] } = useHotspotProfilesQuery(selectedDeviceId || '')
  const { data: users = [] } = useHotspotUsersQuery(selectedDeviceId || '')
  const deleteMutation = useDeleteHotspotUsersMutation()

  // Collect unique comments/batches
  const uniqueComments = Array.from(
    new Set(users.map((u) => u.comment?.trim()).filter(Boolean))
  )

  const handleBulkDelete = async () => {
    if (!selectedDeviceId) return

    let mode = activeTab
    let value = ''

    if (activeTab === 'profile') {
      if (!selectedProfile) {
        toast.error('Please select a profile to delete.')
        return
      }
      value = selectedProfile
    } else if (activeTab === 'comment') {
      const commentToUse = selectedComment === '__custom__' ? customComment : selectedComment
      if (!commentToUse) {
        toast.error('Please select or type a batch tag comment.')
        return
      }
      value = commentToUse
    }

    onOpenChange(false)

    toast.promise(
      deleteMutation.mutateAsync({
        deviceId: selectedDeviceId,
        mode,
        value,
      }),
      {
        loading: `Deleting users (${activeTab})...`,
        success: (res) => `${res.deletedCount} user(s) removed successfully.`,
        error: 'Failed to delete users.',
      }
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[500px]'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2 text-destructive'>
            <Trash2 className='size-5' />
            Bulk User Cleaner
          </DialogTitle>
          <DialogDescription>
            Purge users from MikroTik Hotspot in bulk by profile, batch tag, or expiration.
          </DialogDescription>
        </DialogHeader>

        <Tabs
          value={activeTab}
          onValueChange={(val) => setActiveTab(val as typeof activeTab)}
          className='w-full'
        >
          <TabsList className='grid grid-cols-3 w-full'>
            <TabsTrigger value='profile' className='gap-1.5 text-xs'>
              <UserX className='size-3.5' />
              By Profile
            </TabsTrigger>
            <TabsTrigger value='comment' className='gap-1.5 text-xs'>
              <Tag className='size-3.5' />
              By Comment
            </TabsTrigger>
            <TabsTrigger value='expired' className='gap-1.5 text-xs'>
              <Clock className='size-3.5' />
              Expired Users
            </TabsTrigger>
          </TabsList>

          <div className='py-4 space-y-4'>
            <TabsContent value='profile' className='space-y-3 m-0'>
              <p className='text-xs text-muted-foreground'>
                Delete all hotspot users and vouchers assigned to a specific user profile.
              </p>
              <div className='space-y-1.5'>
                <Label className='text-xs'>Select Profile</Label>
                <Select value={selectedProfile} onValueChange={setSelectedProfile}>
                  <SelectTrigger>
                    <SelectValue placeholder='Choose a profile...' />
                  </SelectTrigger>
                  <SelectContent>
                    {profiles.map((p) => (
                      <SelectItem key={p.id} value={p.name}>
                        {p.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </TabsContent>

            <TabsContent value='comment' className='space-y-3 m-0'>
              <p className='text-xs text-muted-foreground'>
                Delete all vouchers generated under a specific batch comment tag.
              </p>
              <div className='space-y-1.5'>
                <Label className='text-xs'>Select Batch Tag</Label>
                <Select value={selectedComment} onValueChange={setSelectedComment}>
                  <SelectTrigger>
                    <SelectValue placeholder='Choose batch comment tag...' />
                  </SelectTrigger>
                  <SelectContent>
                    {uniqueComments.map((c) => (
                      <SelectItem key={c} value={c}>
                        {c}
                      </SelectItem>
                    ))}
                    <SelectItem value='__custom__'>+ Custom Tag</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              {selectedComment === '__custom__' && (
                <div className='space-y-1.5'>
                  <Label className='text-xs'>Custom Comment Tag</Label>
                  <Input
                    placeholder='e.g. vc-20260818'
                    value={customComment}
                    onChange={(e) => setCustomComment(e.target.value)}
                  />
                </div>
              )}
            </TabsContent>

            <TabsContent value='expired' className='space-y-3 m-0'>
              <p className='text-xs text-muted-foreground'>
                Purge all hotspot users whose uptime has reached their limit or whose validity has expired.
              </p>
              <div className='p-3 bg-muted rounded-md text-xs space-y-1'>
                <div className='font-semibold text-foreground'>Targets:</div>
                <ul className='list-disc pl-4 space-y-0.5 text-muted-foreground'>
                  <li>Users with uptime equal to limit-uptime</li>
                  <li>Users marked with 1s limit-uptime by expire monitor</li>
                  <li>Users tagged with expired comments</li>
                </ul>
              </div>
            </TabsContent>

            <Alert variant='destructive'>
              <AlertTriangle className='size-4' />
              <AlertTitle className='text-xs font-semibold'>Warning</AlertTitle>
              <AlertDescription className='text-xs'>
                This will permanently delete the matching users directly from the active router.
              </AlertDescription>
            </Alert>
          </div>
        </Tabs>

        <DialogFooter className='gap-2'>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant='destructive'
            onClick={handleBulkDelete}
            disabled={deleteMutation.isPending}
            className='gap-1.5'
          >
            <Trash2 className='size-4' />
            {deleteMutation.isPending ? 'Deleting...' : 'Delete Users'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
