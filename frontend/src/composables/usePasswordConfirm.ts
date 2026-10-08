import { reactive } from 'vue'

type Runner = (password: string) => Promise<unknown>

export const passwordConfirmState = reactive({
  open: false,
  title: '',
  message: '',
  run: null as Runner | null,
  resolve: null as ((v: unknown) => void) | null,
})

/**
 * 弹出密码确认框，用输入的密码执行 run；密码错误时停留在弹窗内提示。
 * 成功返回 run 的结果，取消返回 null。密码不做任何缓存，每次调用都需重新输入。
 */
export function confirmWithPassword<T>(
  opts: { title: string; message?: string },
  run: (password: string) => Promise<T>,
): Promise<T | null> {
  passwordConfirmState.resolve?.(null)
  return new Promise<T | null>((resolve) => {
    passwordConfirmState.title = opts.title
    passwordConfirmState.message = opts.message || ''
    passwordConfirmState.run = run
    passwordConfirmState.resolve = resolve as (v: unknown) => void
    passwordConfirmState.open = true
  })
}
