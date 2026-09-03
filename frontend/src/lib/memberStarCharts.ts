import { starapp, type StarChart } from '../api/client'
import { canViewChoresFromStatus, type StatusLike } from './rbacAccess'

export async function listMemberStarCharts(status: StatusLike | null | undefined): Promise<StarChart[]> {
  if (!canViewChoresFromStatus(status)) {
    return []
  }
  try {
    const res = await starapp.listStarCharts({ assignedToMe: true })
    return (res.starCharts || []).filter((chart) => chart.active !== false)
  } catch {
    return []
  }
}
