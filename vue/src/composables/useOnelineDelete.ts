import { showConfirm } from '@common/ui/confirm/confirm'
import { showToast } from '@common/ui/toast/toast'
import httpy from '@common/utils/httpy'

import useAuthStore from '@/stores/auth'

export interface OnelineRow {
  id: number
  user_id: number
}

interface UseOnelineDeleteOptions {
  onSuccess?: (row: OnelineRow) => void | Promise<void>
}

export const useOnelineDelete = (options: UseOnelineDeleteOptions = {}) => {
  const auth = useAuthStore()

  const del = async (row: OnelineRow) => {
    if (!auth.canDelete(row.user_id)) return
    const ok = await showConfirm('이 한줄잡담을 삭제하시겠습니까?')
    if (!ok) return

    const [, err] = await httpy.delete(`/api/onelines/${row.id}`)
    if (err) {
      console.error(err)
      showToast(err.message || '삭제 실패')
      return
    }

    if (options.onSuccess) await options.onSuccess(row)
    showToast('삭제 완료')
  }

  return { del, canDelete: auth.canDelete }
}
