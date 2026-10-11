import type { LocaleBundle } from '../types'

const zh = {
  delete: '删除',
  deleteFailed: '删除失败',
  name: '名称',
  sort: '排序',
  dangerZone: '危险操作',
  deleteDialog: {
    title: '确认删除',
    titleAgain: '再次确认',
    ask: '确定删除「{name}」吗？',
    askAgain: '删除后不可恢复。请再次确认要删除「{name}」。',
    continue: '继续删除',
    confirm: '确认删除',
  },
  defaultAccount: {
    expense: '默认消费',
    repay: '默认还款',
    both: '默认消费 · 默认还款',
  },
  date: {
    pickDate: '选择日期',
    pickTime: '选择时间',
    pickDateTime: '选择日期时间',
    hour: '时',
    minute: '分',
    clear: '清除',
    today: '今天',
    ok: '确定',
    dateFormat: 'YYYY年M月D日',
    dateTimeFormat: 'YYYY年M月D日 HH:mm',
  },
}

const en: typeof zh = {
  delete: 'Delete',
  deleteFailed: 'Failed to delete',
  name: 'Name',
  sort: 'Sort order',
  dangerZone: 'Danger zone',
  deleteDialog: {
    title: 'Confirm deletion',
    titleAgain: 'Confirm again',
    ask: 'Delete "{name}"?',
    askAgain: 'This cannot be undone. Please confirm again that you want to delete "{name}".',
    continue: 'Continue',
    confirm: 'Delete',
  },
  defaultAccount: {
    expense: 'Default for spending',
    repay: 'Default for repayment',
    both: 'Default for spending · repayment',
  },
  date: {
    pickDate: 'Select date',
    pickTime: 'Select time',
    pickDateTime: 'Select date & time',
    hour: 'Hour',
    minute: 'Minute',
    clear: 'Clear',
    today: 'Today',
    ok: 'OK',
    dateFormat: 'MMM D, YYYY',
    dateTimeFormat: 'MMM D, YYYY HH:mm',
  },
}

const ja: typeof zh = {
  delete: '削除',
  deleteFailed: '削除に失敗しました',
  name: '名前',
  sort: '並び順',
  dangerZone: '危険な操作',
  deleteDialog: {
    title: '削除の確認',
    titleAgain: '再確認',
    ask: '「{name}」を削除しますか？',
    askAgain: '削除すると元に戻せません。「{name}」を削除してよいか、もう一度確認してください。',
    continue: '削除を続ける',
    confirm: '削除する',
  },
  defaultAccount: {
    expense: '既定の支払い',
    repay: '既定の返済',
    both: '既定の支払い · 既定の返済',
  },
  date: {
    pickDate: '日付を選択',
    pickTime: '時刻を選択',
    pickDateTime: '日時を選択',
    hour: '時',
    minute: '分',
    clear: 'クリア',
    today: '今日',
    ok: 'OK',
    dateFormat: 'YYYY年M月D日',
    dateTimeFormat: 'YYYY年M月D日 HH:mm',
  },
}

const ko: typeof zh = {
  delete: '삭제',
  deleteFailed: '삭제에 실패했습니다',
  name: '이름',
  sort: '정렬 순서',
  dangerZone: '위험 작업',
  deleteDialog: {
    title: '삭제 확인',
    titleAgain: '다시 확인',
    ask: '「{name}」을(를) 삭제하시겠습니까?',
    askAgain: '삭제 후에는 복구할 수 없습니다. 「{name}」을(를) 삭제할지 다시 확인하세요.',
    continue: '계속 삭제',
    confirm: '삭제',
  },
  defaultAccount: {
    expense: '기본 결제',
    repay: '기본 상환',
    both: '기본 결제 · 기본 상환',
  },
  date: {
    pickDate: '날짜 선택',
    pickTime: '시간 선택',
    pickDateTime: '날짜 및 시간 선택',
    hour: '시',
    minute: '분',
    clear: '지우기',
    today: '오늘',
    ok: '확인',
    dateFormat: 'YYYY년 M월 D일',
    dateTimeFormat: 'YYYY년 M월 D일 HH:mm',
  },
}

export default { 'zh-CN': zh, en, ja, ko } satisfies LocaleBundle<typeof zh>
