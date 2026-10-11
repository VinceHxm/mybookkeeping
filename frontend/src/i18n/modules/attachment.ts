import type { LocaleBundle } from '../types'

const zh = {
  title: '附件',
  subtitle: '票据 / 小票照片，可选 · 点图可放大',
  count: '{n} 张',
  addImage: '添加图片',
  addImageSub: '相册或文件',
  takePhoto: '拍照',
  takePhotoSub: '手机摄像头',
  preview: '附件预览',
  remove: '删除附件',
  addMore: '继续加',
  closePreview: '关闭预览',
  prev: '上一张',
  next: '下一张',
  large: '附件大图',
  uploadFailed: '上传失败',
}

const en: typeof zh = {
  title: 'Attachments',
  subtitle: 'Receipt photos, optional · tap to enlarge',
  count: '{n} photo(s)',
  addImage: 'Add image',
  addImageSub: 'Album or file',
  takePhoto: 'Take photo',
  takePhotoSub: 'Phone camera',
  preview: 'Attachment preview',
  remove: 'Remove attachment',
  addMore: 'Add',
  closePreview: 'Close preview',
  prev: 'Previous',
  next: 'Next',
  large: 'Attachment full size',
  uploadFailed: 'Upload failed',
}

const ja: typeof zh = {
  title: '添付ファイル',
  subtitle: 'レシートの写真（任意）· タップで拡大',
  count: '{n} 枚',
  addImage: '画像を追加',
  addImageSub: 'アルバムまたはファイル',
  takePhoto: '撮影',
  takePhotoSub: 'スマホのカメラ',
  preview: '添付ファイルのプレビュー',
  remove: '添付ファイルを削除',
  addMore: '追加',
  closePreview: 'プレビューを閉じる',
  prev: '前へ',
  next: '次へ',
  large: '添付画像（拡大）',
  uploadFailed: 'アップロードに失敗しました',
}

const ko: typeof zh = {
  title: '첨부 파일',
  subtitle: '영수증 사진, 선택 · 탭하면 확대',
  count: '{n}장',
  addImage: '이미지 추가',
  addImageSub: '앨범 또는 파일',
  takePhoto: '사진 촬영',
  takePhotoSub: '휴대폰 카메라',
  preview: '첨부 파일 미리보기',
  remove: '첨부 파일 삭제',
  addMore: '추가',
  closePreview: '미리보기 닫기',
  prev: '이전',
  next: '다음',
  large: '첨부 이미지 크게 보기',
  uploadFailed: '업로드에 실패했습니다',
}

export default { 'zh-CN': zh, en, ja, ko } satisfies LocaleBundle<typeof zh>
