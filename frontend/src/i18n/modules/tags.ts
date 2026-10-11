import type { LocaleBundle } from '../types'

const zh = {
  title: '标签管理',
  meta: '标签色用于流水筛选与展示',
  empty: '暂无标签',
  edit: '编辑标签',
  create: '新建标签',
  info: '标签信息',
  color: '颜色',
  deleteTip: '删除标签不可恢复，已关联流水会失去该标签。',
  deleteThis: '删除此标签',
}

const en: typeof zh = {
  title: 'Tags',
  meta: 'Tag colors are used for filtering and displaying records',
  empty: 'No tags yet',
  edit: 'Edit tag',
  create: 'New tag',
  info: 'Tag info',
  color: 'Color',
  deleteTip: 'Deleting a tag cannot be undone. Linked records will lose this tag.',
  deleteThis: 'Delete this tag',
}

const ja: typeof zh = {
  title: 'タグ管理',
  meta: 'タグの色は記録の絞り込みと表示に使われます',
  empty: 'タグはまだありません',
  edit: 'タグを編集',
  create: 'タグを作成',
  info: 'タグ情報',
  color: '色',
  deleteTip: 'タグの削除は元に戻せません。関連付けられた記録からこのタグが外れます。',
  deleteThis: 'このタグを削除',
}

const ko: typeof zh = {
  title: '태그 관리',
  meta: '태그 색상은 내역 필터링과 표시에 사용됩니다',
  empty: '태그가 없습니다',
  edit: '태그 편집',
  create: '새 태그',
  info: '태그 정보',
  color: '색상',
  deleteTip: '태그를 삭제하면 복구할 수 없으며, 연결된 내역에서 이 태그가 제거됩니다.',
  deleteThis: '이 태그 삭제',
}

export default { 'zh-CN': zh, en, ja, ko } satisfies LocaleBundle<typeof zh>
