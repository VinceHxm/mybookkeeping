/** 是否为脱敏展示串（含 · / *），不可当作明文写回 */
export function looksMaskedSecret(raw?: string | null): boolean {
  const s = (raw || '').trim()
  if (!s) return false
  return /[·•*＊]/.test(s)
}

/** 仅用于展示已脱敏的卡号；明文不应出现在前端常驻状态 */
export function displayCardNo(masked?: string | null): string {
  return (masked || '').trim()
}

export function displayHolderName(masked?: string | null): string {
  return (masked || '').trim()
}

/**
 * 保存时：若仍是脱敏串则不下发该字段（后端也会忽略）；
 * 空串表示清空；明文表示更新。
 */
export function secretFieldForSave(
  current: string,
  originalMasked: string,
): string | undefined {
  const v = (current || '').trim()
  const orig = (originalMasked || '').trim()
  if (looksMaskedSecret(v)) return undefined
  if (v === orig && looksMaskedSecret(orig)) return undefined
  // 未改动且本来就是空
  if (v === orig) return undefined
  return v
}
