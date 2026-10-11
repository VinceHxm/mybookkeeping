import type { LocaleBundle } from '../types'

const zh = {
  title: '分类管理',
  root: '一级分类',
  childCount: '{n} 个子分类',
  belongsTo: '属于「{name}」',
  emptyExpense: '暂无支出分类',
  emptyIncome: '暂无收入分类',
  edit: '编辑分类',
  create: '新建分类',
  thisCategory: '此分类',
  form: {
    basicInfo: '基本信息',
    icon: '分类图标',
    type: '类型',
    parent: '父分类（可选）',
    noParent: '无（一级分类）',
    hasChildrenTip: '该分类下已有子分类，不可再挂到其他父分类下。',
    deleteTip: '删除分类不可恢复；若有子分类或流水将无法删除。',
    deleteThis: '删除此分类',
  },
}

const en: typeof zh = {
  title: 'Categories',
  root: 'Parent category',
  childCount: '{n} sub category(s)',
  belongsTo: 'Under "{name}"',
  emptyExpense: 'No expense categories yet',
  emptyIncome: 'No income categories yet',
  edit: 'Edit category',
  create: 'New category',
  thisCategory: 'this category',
  form: {
    basicInfo: 'Basic info',
    icon: 'Category icon',
    type: 'Type',
    parent: 'Parent category (optional)',
    noParent: 'None (parent category)',
    hasChildrenTip: 'This category has sub categories, so it cannot be moved under another parent.',
    deleteTip: 'Deleting a category cannot be undone. Categories with sub categories or records cannot be deleted.',
    deleteThis: 'Delete this category',
  },
}

const ja: typeof zh = {
  title: 'カテゴリ管理',
  root: '大分類',
  childCount: '小分類 {n} 件',
  belongsTo: '「{name}」の小分類',
  emptyExpense: '支出カテゴリはまだありません',
  emptyIncome: '収入カテゴリはまだありません',
  edit: 'カテゴリを編集',
  create: 'カテゴリを作成',
  thisCategory: 'このカテゴリ',
  form: {
    basicInfo: '基本情報',
    icon: 'カテゴリアイコン',
    type: '種類',
    parent: '大分類（任意）',
    noParent: 'なし（大分類にする）',
    hasChildrenTip: 'このカテゴリには小分類があるため、他の大分類の下には移動できません。',
    deleteTip: 'カテゴリの削除は元に戻せません。小分類や記録がある場合は削除できません。',
    deleteThis: 'このカテゴリを削除',
  },
}

const ko: typeof zh = {
  title: '분류 관리',
  root: '대분류',
  childCount: '소분류 {n}개',
  belongsTo: '「{name}」의 소분류',
  emptyExpense: '지출 분류가 없습니다',
  emptyIncome: '수입 분류가 없습니다',
  edit: '분류 편집',
  create: '새 분류',
  thisCategory: '이 분류',
  form: {
    basicInfo: '기본 정보',
    icon: '분류 아이콘',
    type: '유형',
    parent: '대분류 (선택)',
    noParent: '없음 (대분류로 지정)',
    hasChildrenTip: '이 분류에는 소분류가 있어 다른 대분류 아래로 옮길 수 없습니다.',
    deleteTip: '분류를 삭제하면 복구할 수 없습니다. 소분류나 내역이 있으면 삭제할 수 없습니다.',
    deleteThis: '이 분류 삭제',
  },
}

export default { 'zh-CN': zh, en, ja, ko } satisfies LocaleBundle<typeof zh>
