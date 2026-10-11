import type { LocaleBundle } from '../types'

const zh = {
  title: '校准市值',
  editTitle: '编辑市值校准',
  bookBalance: '账上余额 ¥{amount}',
  marketValue: '当前市值（元）',
  same: '与账上一致，只记录校准时间',
  gain: '本次浮盈 {amount}',
  loss: '本次浮亏 {amount}',
  delta: '校准差额（元，浮亏填负数）',
  time: '校准时间',
  remark: '备注（可选）',
  hint: '照着支付宝 / 券商 App 的持仓市值填。差额记为一笔「市值校准」，只改这个账户的余额，不计入收支统计。',
  delete: '删除',
  deltaZero: '差额不能为 0，不需要这笔校准可直接删除',
  needMarketValue: '请填写当前市值',
  saveFailed: '保存失败',
  deleteFailed: '删除失败',
}

const en: typeof zh = {
  title: 'Revalue',
  editTitle: 'Edit revaluation',
  bookBalance: 'Book balance ¥{amount}',
  marketValue: 'Current market value (¥)',
  same: 'Matches the book balance; only the revaluation time is recorded',
  gain: 'Unrealized gain {amount}',
  loss: 'Unrealized loss {amount}',
  delta: 'Revaluation difference (¥, negative for a loss)',
  time: 'Revaluation time',
  remark: 'Note (optional)',
  hint: 'Enter the holding value shown in Alipay or your broker app. The difference is recorded as a "Revaluation" that only changes this account\'s balance and is excluded from income / expense stats.',
  delete: 'Delete',
  deltaZero: 'The difference cannot be 0. Delete this revaluation if you no longer need it',
  needMarketValue: 'Please enter the current market value',
  saveFailed: 'Failed to save',
  deleteFailed: 'Failed to delete',
}

const ja: typeof zh = {
  title: '評価額を調整',
  editTitle: '評価額調整を編集',
  bookBalance: '帳簿上の残高 ¥{amount}',
  marketValue: '現在の評価額（元）',
  same: '帳簿と一致しています。調整日時のみ記録します',
  gain: '今回の含み益 {amount}',
  loss: '今回の含み損 {amount}',
  delta: '調整差額（元、含み損はマイナスで入力）',
  time: '調整日時',
  remark: 'メモ（任意）',
  hint: 'Alipay や証券アプリの保有評価額を入力してください。差額は「評価額調整」として記録され、この口座の残高だけを変更し、収支の集計には含まれません。',
  delete: '削除',
  deltaZero: '差額を 0 にはできません。この調整が不要なら削除してください',
  needMarketValue: '現在の評価額を入力してください',
  saveFailed: '保存に失敗しました',
  deleteFailed: '削除に失敗しました',
}

const ko: typeof zh = {
  title: '평가액 조정',
  editTitle: '평가액 조정 편집',
  bookBalance: '장부 잔액 ¥{amount}',
  marketValue: '현재 평가액(위안)',
  same: '장부와 일치합니다. 조정 시간만 기록합니다',
  gain: '이번 평가 이익 {amount}',
  loss: '이번 평가 손실 {amount}',
  delta: '조정 차액(위안, 손실은 음수로 입력)',
  time: '조정 시간',
  remark: '메모(선택)',
  hint: '알리페이나 증권사 앱의 보유 평가액을 입력하세요. 차액은 「평가액 조정」으로 기록되며, 이 계좌의 잔액만 바뀌고 수입/지출 통계에는 포함되지 않습니다.',
  delete: '삭제',
  deltaZero: '차액은 0일 수 없습니다. 이 조정이 필요 없으면 삭제하세요',
  needMarketValue: '현재 평가액을 입력하세요',
  saveFailed: '저장하지 못했습니다',
  deleteFailed: '삭제하지 못했습니다',
}

export default { 'zh-CN': zh, en, ja, ko } satisfies LocaleBundle<typeof zh>
