import type { LocaleBundle } from '../types'

const zh = {
  title: '添加到主屏幕，像 App 一样使用',
  body: '点击浏览器的 {share} 分享按钮，再选择「添加到主屏幕」',
  ok: '知道了',
}

const en: typeof zh = {
  title: 'Add to Home Screen for an app-like experience',
  body: 'Tap the {share} Share button, then choose "Add to Home Screen"',
  ok: 'Got it',
}

const ja: typeof zh = {
  title: 'ホーム画面に追加してアプリのように使う',
  body: 'ブラウザの {share} 共有ボタンをタップし、「ホーム画面に追加」を選択してください',
  ok: 'OK',
}

const ko: typeof zh = {
  title: '홈 화면에 추가해 앱처럼 사용하기',
  body: '브라우저의 {share} 공유 버튼을 누른 뒤 "홈 화면에 추가"를 선택하세요',
  ok: '확인',
}

export default { 'zh-CN': zh, en, ja, ko } satisfies LocaleBundle<typeof zh>
