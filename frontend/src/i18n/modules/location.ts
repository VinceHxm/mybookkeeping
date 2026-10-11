import type { LocaleBundle } from '../types'

const zh = {
  title: '地点',
  subtitle: '消费位置或路费起终点，可选',
  route: '行程',
  point: '位置',
  add: '添加地点',
  addSub: '点选地图，或一键定位当前位置',
  pickedRoute: '已选行程',
  pickedPoint: '已选位置',
  edit: '修改',
  clear: '清除',
}

const en: typeof zh = {
  title: 'Location',
  subtitle: 'Where you spent, or trip start and end — optional',
  route: 'Trip',
  point: 'Place',
  add: 'Add location',
  addSub: 'Pick on the map, or locate yourself in one tap',
  pickedRoute: 'Selected trip',
  pickedPoint: 'Selected place',
  edit: 'Edit',
  clear: 'Clear',
}

const ja: typeof zh = {
  title: '場所',
  subtitle: '利用した場所、または交通費の出発地と到着地（任意）',
  route: '経路',
  point: '地点',
  add: '場所を追加',
  addSub: '地図で選ぶか、ワンタップで現在地を取得',
  pickedRoute: '選択した経路',
  pickedPoint: '選択した地点',
  edit: '変更',
  clear: 'クリア',
}

const ko: typeof zh = {
  title: '위치',
  subtitle: '결제 위치 또는 교통비 출발지와 도착지, 선택',
  route: '경로',
  point: '위치',
  add: '위치 추가',
  addSub: '지도에서 선택하거나 한 번에 현재 위치 찾기',
  pickedRoute: '선택한 경로',
  pickedPoint: '선택한 위치',
  edit: '수정',
  clear: '지우기',
}

export default { 'zh-CN': zh, en, ja, ko } satisfies LocaleBundle<typeof zh>
