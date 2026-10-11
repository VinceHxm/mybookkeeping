import type { LocaleBundle } from '../types'

const zh = {
  title: '外币支付',
  subtitle: '记录外币原价和实际人民币扣款，自动算支付汇率',
  currency: '币种',
  amount: '外币金额',
  estRate: '预估汇率',
  payRate: '支付汇率',
  rateLine: '{label}：1 {code} = {rate} CNY',
  rateHint: '填好外币金额与上方人民币金额后，自动算出支付汇率',
}

const en: typeof zh = {
  title: 'Foreign currency',
  subtitle: 'Record the foreign price and the actual CNY charged to get the exchange rate',
  currency: 'Currency',
  amount: 'Foreign amount',
  estRate: 'Estimated rate',
  payRate: 'Exchange rate',
  rateLine: '{label}: 1 {code} = {rate} CNY',
  rateHint: 'Enter the foreign amount and the CNY amount above to calculate the exchange rate',
}

const ja: typeof zh = {
  title: '外貨払い',
  subtitle: '外貨の価格と実際の人民元の引き落とし額を記録し、為替レートを自動計算します',
  currency: '通貨',
  amount: '外貨金額',
  estRate: '見込みレート',
  payRate: '支払いレート',
  rateLine: '{label}：1 {code} = {rate} CNY',
  rateHint: '外貨金額と上の人民元金額を入力すると、支払いレートが自動計算されます',
}

const ko: typeof zh = {
  title: '외화 결제',
  subtitle: '외화 원가와 실제 위안화 결제액을 기록해 환율을 자동 계산합니다',
  currency: '통화',
  amount: '외화 금액',
  estRate: '예상 환율',
  payRate: '결제 환율',
  rateLine: '{label}: 1 {code} = {rate} CNY',
  rateHint: '외화 금액과 위의 위안화 금액을 입력하면 결제 환율이 자동 계산됩니다',
}

export default { 'zh-CN': zh, en, ja, ko } satisfies LocaleBundle<typeof zh>
