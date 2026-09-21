/** 记账常用 MDI 图标库（项目已引入 @mdi/font） */
export type IconItem = { value: string; label: string; group: string }

export const ICON_GROUPS = [
  '常用',
  '餐饮',
  '交通',
  '购物生活',
  '居住',
  '娱乐社交',
  '医疗教育',
  '收入理财',
  '标记',
] as const

export const APP_ICONS: IconItem[] = [
  // 常用
  { value: 'mdi-shape', label: '形状', group: '常用' },
  { value: 'mdi-star', label: '星标', group: '常用' },
  { value: 'mdi-heart', label: '爱心', group: '常用' },
  { value: 'mdi-bookmark', label: '书签', group: '常用' },
  { value: 'mdi-flash', label: '闪电', group: '常用' },
  { value: 'mdi-checkbox-marked-circle', label: '完成', group: '常用' },
  { value: 'mdi-dots-horizontal', label: '其他', group: '常用' },

  // 餐饮
  { value: 'mdi-food', label: '餐饮', group: '餐饮' },
  { value: 'mdi-food-apple', label: '水果', group: '餐饮' },
  { value: 'mdi-coffee', label: '咖啡', group: '餐饮' },
  { value: 'mdi-beer', label: '酒水', group: '餐饮' },
  { value: 'mdi-noodles', label: '面食', group: '餐饮' },
  { value: 'mdi-pizza', label: '披萨', group: '餐饮' },
  { value: 'mdi-cup', label: '饮品', group: '餐饮' },

  // 交通
  { value: 'mdi-bus', label: '公交', group: '交通' },
  { value: 'mdi-subway-variant', label: '地铁', group: '交通' },
  { value: 'mdi-train', label: '火车', group: '交通' },
  { value: 'mdi-airplane', label: '飞机', group: '交通' },
  { value: 'mdi-taxi', label: '打车', group: '交通' },
  { value: 'mdi-car', label: '汽车', group: '交通' },
  { value: 'mdi-bike', label: '骑行', group: '交通' },
  { value: 'mdi-gas-station', label: '加油', group: '交通' },
  { value: 'mdi-parking', label: '停车', group: '交通' },
  { value: 'mdi-map-marker', label: '地点', group: '交通' },
  { value: 'mdi-map-marker-path', label: '行程', group: '交通' },
  { value: 'mdi-ticket-percent', label: '票卡', group: '交通' },

  // 购物生活
  { value: 'mdi-cart', label: '购物', group: '购物生活' },
  { value: 'mdi-shopping', label: '购物袋', group: '购物生活' },
  { value: 'mdi-store', label: '商店', group: '购物生活' },
  { value: 'mdi-gift', label: '礼物', group: '购物生活' },
  { value: 'mdi-tshirt-crew', label: '服饰', group: '购物生活' },
  { value: 'mdi-shoe-formal', label: '鞋靴', group: '购物生活' },
  { value: 'mdi-cellphone', label: '手机', group: '购物生活' },
  { value: 'mdi-laptop', label: '数码', group: '购物生活' },
  { value: 'mdi-scissors-cutting', label: '美容', group: '购物生活' },
  { value: 'mdi-dog', label: '宠物', group: '购物生活' },

  // 居住
  { value: 'mdi-home', label: '居住', group: '居住' },
  { value: 'mdi-home-city', label: '房产', group: '居住' },
  { value: 'mdi-lightning-bolt', label: '水电', group: '居住' },
  { value: 'mdi-water', label: '水费', group: '居住' },
  { value: 'mdi-wifi', label: '网络', group: '居住' },
  { value: 'mdi-sofa', label: '家居', group: '居住' },
  { value: 'mdi-broom', label: '家政', group: '居住' },
  { value: 'mdi-tools', label: '维修', group: '居住' },

  // 娱乐社交
  { value: 'mdi-movie', label: '影视', group: '娱乐社交' },
  { value: 'mdi-gamepad-variant', label: '游戏', group: '娱乐社交' },
  { value: 'mdi-music', label: '音乐', group: '娱乐社交' },
  { value: 'mdi-microphone', label: 'K歌', group: '娱乐社交' },
  { value: 'mdi-account-group', label: '社交', group: '娱乐社交' },
  { value: 'mdi-beach', label: '旅行', group: '娱乐社交' },
  { value: 'mdi-camera', label: '摄影', group: '娱乐社交' },
  { value: 'mdi-book-open-page-variant', label: '阅读', group: '娱乐社交' },

  // 医疗教育
  { value: 'mdi-hospital', label: '医疗', group: '医疗教育' },
  { value: 'mdi-pill', label: '药品', group: '医疗教育' },
  { value: 'mdi-tooth', label: '牙科', group: '医疗教育' },
  { value: 'mdi-school', label: '教育', group: '医疗教育' },
  { value: 'mdi-bookshelf', label: '学习', group: '医疗教育' },
  { value: 'mdi-baby-carriage', label: '育儿', group: '医疗教育' },
  { value: 'mdi-dumbbell', label: '健身', group: '医疗教育' },

  // 收入理财
  { value: 'mdi-cash', label: '现金', group: '收入理财' },
  { value: 'mdi-cash-multiple', label: '工资', group: '收入理财' },
  { value: 'mdi-wallet', label: '钱包', group: '收入理财' },
  { value: 'mdi-credit-card', label: '信用卡', group: '收入理财' },
  { value: 'mdi-bank', label: '银行', group: '收入理财' },
  { value: 'mdi-chart-line', label: '投资', group: '收入理财' },
  { value: 'mdi-piggy-bank', label: '储蓄', group: '收入理财' },
  { value: 'mdi-handshake', label: '兼职', group: '收入理财' },
  { value: 'mdi-trophy', label: '奖金', group: '收入理财' },
  { value: 'mdi-receipt', label: '账单', group: '收入理财' },
  { value: 'mdi-calculator', label: '计算', group: '收入理财' },
  { value: 'mdi-calendar-clock', label: '周期', group: '收入理财' },

  // 标记
  { value: 'mdi-tag', label: '标签', group: '标记' },
  { value: 'mdi-flag', label: '旗帜', group: '标记' },
  { value: 'mdi-alert-circle', label: '提醒', group: '标记' },
  { value: 'mdi-information', label: '信息', group: '标记' },
  { value: 'mdi-fire', label: '热点', group: '标记' },
  { value: 'mdi-leaf', label: '叶子', group: '标记' },
  { value: 'mdi-flower', label: '花朵', group: '标记' },
  { value: 'mdi-weather-sunny', label: '晴天', group: '标记' },
]

export function iconLabel(value: string): string {
  return APP_ICONS.find((i) => i.value === value)?.label || value.replace(/^mdi-/, '')
}

export function accountTypeIcon(type: string): string {
  switch (type) {
    case 'cash': return 'mdi-cash'
    case 'bank': return 'mdi-bank'
    case 'credit': return 'mdi-credit-card'
    default: return 'mdi-wallet'
  }
}

export function txTypeIcon(type: string): string {
  switch (type) {
    case 'income': return 'mdi-arrow-down-bold-circle'
    case 'transfer': return 'mdi-swap-horizontal'
    default: return 'mdi-arrow-up-bold-circle'
  }
}
