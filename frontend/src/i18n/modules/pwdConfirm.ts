import type { LocaleBundle } from '../types'

const zh = {
  password: '登录密码',
  confirm: '确认查看',
  failed: '验证失败',
  verifyTitle: '验证身份',
  revealMessage: '查看完整卡号与户名前，请输入登录密码。',
}

const en: typeof zh = {
  password: 'Password',
  confirm: 'View',
  failed: 'Verification failed',
  verifyTitle: 'Verify your identity',
  revealMessage: 'Enter your password to view the full card number and holder name.',
}

const ja: typeof zh = {
  password: 'ログインパスワード',
  confirm: '表示する',
  failed: '認証に失敗しました',
  verifyTitle: '本人確認',
  revealMessage: 'カード番号と名義人をすべて表示するには、ログインパスワードを入力してください。',
}

const ko: typeof zh = {
  password: '로그인 비밀번호',
  confirm: '확인하고 보기',
  failed: '인증에 실패했습니다',
  verifyTitle: '본인 확인',
  revealMessage: '전체 카드 번호와 예금주를 보려면 로그인 비밀번호를 입력하세요.',
}

export default { 'zh-CN': zh, en, ja, ko } satisfies LocaleBundle<typeof zh>
