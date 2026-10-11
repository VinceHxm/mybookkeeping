import type { LocaleBundle } from '../types'

const zh = {
  title: '重置密码',
  hint: '默认不开放自助找回，请联系管理员在「用户管理」中为你重置密码。',
  stepToken: '获取令牌',
  stepPassword: '设置新密码',
  generate: '生成重置令牌',
  tokenShown: '令牌：{token}（已自动填入下一步）',
  noToken: '未获取到重置令牌，请联系管理员重置密码',
  token: '重置令牌',
  newPassword: '新密码',
  confirm: '确认重置',
  done: '密码已重置，请登录',
  backToLogin: '返回登录',
}

const en: typeof zh = {
  title: 'Reset password',
  hint: 'Self-service recovery is off by default. Ask an admin to reset your password in Users.',
  stepToken: 'Get token',
  stepPassword: 'Set new password',
  generate: 'Generate reset token',
  tokenShown: 'Token: {token} (filled in for the next step)',
  noToken: 'No reset token was issued. Please ask an admin to reset your password',
  token: 'Reset token',
  newPassword: 'New password',
  confirm: 'Reset password',
  done: 'Password reset. Please sign in',
  backToLogin: 'Back to sign in',
}

const ja: typeof zh = {
  title: 'パスワードリセット',
  hint: '既定ではセルフサービスでの再設定はできません。管理者に「ユーザー管理」からパスワードをリセットしてもらってください。',
  stepToken: 'トークン取得',
  stepPassword: '新しいパスワード',
  generate: 'リセットトークンを発行',
  tokenShown: 'トークン：{token}（次のステップに自動入力しました）',
  noToken: 'リセットトークンを取得できませんでした。管理者にパスワードのリセットを依頼してください',
  token: 'リセットトークン',
  newPassword: '新しいパスワード',
  confirm: 'リセットする',
  done: 'パスワードをリセットしました。ログインしてください',
  backToLogin: 'ログインに戻る',
}

const ko: typeof zh = {
  title: '비밀번호 초기화',
  hint: '기본적으로 직접 비밀번호를 찾을 수 없습니다. 관리자에게 「사용자 관리」에서 비밀번호 초기화를 요청하세요.',
  stepToken: '토큰 받기',
  stepPassword: '새 비밀번호 설정',
  generate: '초기화 토큰 생성',
  tokenShown: '토큰: {token} (다음 단계에 자동 입력됨)',
  noToken: '초기화 토큰을 받지 못했습니다. 관리자에게 비밀번호 초기화를 요청하세요',
  token: '초기화 토큰',
  newPassword: '새 비밀번호',
  confirm: '초기화',
  done: '비밀번호가 초기화되었습니다. 로그인하세요',
  backToLogin: '로그인으로 돌아가기',
}

export default { 'zh-CN': zh, en, ja, ko } satisfies LocaleBundle<typeof zh>
