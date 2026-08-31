export const MEMBER_STAR_COLOR_PALETTE = [
  '#e74c3c',
  '#3498db',
  '#2ecc71',
  '#f39c12',
  '#9b59b6',
  '#1abc9c',
  '#e67e22',
  '#ff6b6b',
  '#4ecdc4',
  '#6c5ce7',
] as const

export type MemberColorSource = {
  id?: number
  starColor?: string
}

export function memberStarColor(member?: MemberColorSource | null): string {
  if (member?.starColor) return member.starColor
  if (member?.id != null && MEMBER_STAR_COLOR_PALETTE.length) {
    const idx = ((member.id % MEMBER_STAR_COLOR_PALETTE.length) + MEMBER_STAR_COLOR_PALETTE.length) % MEMBER_STAR_COLOR_PALETTE.length
    return MEMBER_STAR_COLOR_PALETTE[idx]
  }
  return MEMBER_STAR_COLOR_PALETTE[0]
}

export function memberAvatarStyle(member?: MemberColorSource | null): Record<string, string> {
  const color = memberStarColor(member)
  return {
    border: `3px solid ${color}`,
    '--member-star-color': color,
  }
}

export function memberStarStyle(member?: MemberColorSource | null): Record<string, string> {
  return { color: memberStarColor(member) }
}
