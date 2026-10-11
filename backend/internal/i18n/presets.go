package i18n

// 默认分类 / 标签名：键为 service 里预设树的中文名，新用户按注册语言写入译名。
var presetNames = map[string]entry{
	// 支出
	"餐饮":   {"Food & Dining", "食費", "식비"},
	"食品":   {"Groceries", "食料品", "식료품"},
	"饮料":   {"Drinks", "飲み物", "음료"},
	"水果零食": {"Fruit & Snacks", "果物・おやつ", "과일·간식"},
	"外卖":   {"Takeout", "デリバリー", "배달음식"},
	"服饰外貌": {"Clothing & Beauty", "衣服・美容", "의류·미용"},
	"衣服":   {"Clothes", "衣服", "의류"},
	"饰品":   {"Accessories", "アクセサリー", "액세서리"},
	"化妆品":  {"Cosmetics", "化粧品", "화장품"},
	"美容美发": {"Beauty & Hair", "美容・ヘアカット", "미용·헤어"},
	"居住":   {"Housing", "住まい", "주거"},
	"家居用品": {"Household Goods", "日用品・家具", "생활용품"},
	"电子产品": {"Electronics", "電子機器", "전자제품"},
	"维修保养": {"Repairs & Maintenance", "修理・メンテナンス", "수리·유지보수"},
	"家政服务": {"Housekeeping", "家事代行", "가사 서비스"},
	"水电煤气": {"Utilities", "水道光熱費", "공과금"},
	"租金房贷": {"Rent & Mortgage", "家賃・住宅ローン", "월세·주택대출"},
	"交通":   {"Transportation", "交通", "교통"},
	"公共交通": {"Public Transit", "公共交通機関", "대중교통"},
	"打车租车": {"Taxi & Car Rental", "タクシー・レンタカー", "택시·렌터카"},
	"私家车":  {"Car", "マイカー", "자가용"},
	"火车票":  {"Train Tickets", "電車・新幹線", "기차표"},
	"飞机票":  {"Flights", "航空券", "항공권"},
	"通讯":   {"Communication", "通信", "통신"},
	"话费":   {"Phone Bill", "携帯料金", "휴대폰 요금"},
	"网费":   {"Internet", "インターネット", "인터넷 요금"},
	"快递费":  {"Shipping", "配送料", "택배비"},
	"购物":   {"Shopping", "買い物", "쇼핑"},
	"日用百货": {"Daily Necessities", "日用雑貨", "생활잡화"},
	"数码电器": {"Digital & Appliances", "デジタル・家電", "디지털·가전"},
	"其他购物": {"Other Shopping", "その他の買い物", "기타 쇼핑"},
	"娱乐":   {"Entertainment", "娯楽", "여가"},
	"运动健身": {"Sports & Fitness", "スポーツ・フィットネス", "운동·헬스"},
	"聚会":   {"Gatherings", "交際費", "모임"},
	"电影演出": {"Movies & Shows", "映画・公演", "영화·공연"},
	"游戏玩具": {"Games & Toys", "ゲーム・おもちゃ", "게임·장난감"},
	"会员订阅": {"Subscriptions", "サブスク", "구독"},
	"宠物":   {"Pets", "ペット", "반려동물"},
	"旅游":   {"Travel", "旅行", "여행"},
	"教育学习": {"Education", "教育・学習", "교육"},
	"书报杂志": {"Books & Magazines", "書籍・雑誌", "도서·잡지"},
	"培训课程": {"Courses", "講座・習い事", "강의·학원"},
	"考试认证": {"Exams & Certificates", "試験・資格", "시험·자격증"},
	"礼物捐赠": {"Gifts & Donations", "贈り物・寄付", "선물·기부"},
	"礼物":   {"Gifts", "贈り物", "선물"},
	"捐赠":   {"Donations", "寄付", "기부"},
	"医疗健康": {"Health & Medical", "医療・健康", "의료·건강"},
	"检查治疗": {"Checkups & Treatment", "診察・治療", "진료·치료"},
	"药品":   {"Medicine", "薬", "의약품"},
	"医疗器械": {"Medical Supplies", "医療用品", "의료용품"},
	"金融保险": {"Finance & Insurance", "金融・保険", "금융·보험"},
	"税费":   {"Taxes", "税金", "세금"},
	"手续费":  {"Fees", "手数料", "수수료"},
	"保险":   {"Insurance", "保険", "보험"},
	"利息支出": {"Interest Paid", "支払利息", "이자 지출"},
	"罚款赔偿": {"Fines & Compensation", "罰金・賠償", "벌금·배상"},
	"其他支出": {"Other Expenses", "その他の支出", "기타 지출"},
	"其他":   {"Other", "その他", "기타"},

	// 收入
	"职业收入": {"Work Income", "給与所得", "근로소득"},
	"工资":   {"Salary", "給与", "급여"},
	"奖金":   {"Bonus", "賞与", "상여금"},
	"加班":   {"Overtime", "残業代", "초과근무수당"},
	"兼职":   {"Side Job", "副業", "부업"},
	"金融投资": {"Investments", "投資", "투자"},
	"投资收入": {"Investment Income", "投資収益", "투자 수익"},
	"租金收入": {"Rental Income", "家賃収入", "임대 수입"},
	"利息收入": {"Interest Income", "受取利息", "이자 수입"},
	"其他收入": {"Other Income", "その他の収入", "기타 수입"},
	"礼品红包": {"Gifts & Red Packets", "お祝い・お年玉", "선물·용돈"},
	"中奖":   {"Winnings", "当選金", "당첨금"},
	"意外收入": {"Windfall", "臨時収入", "뜻밖의 수입"},
	"退款":   {"Refund", "返金", "환불"},

	// 标签
	"报销":  {"Reimbursable", "経費精算", "경비 청구"},
	"出差":  {"Business Trip", "出張", "출장"},
	"工作":  {"Work", "仕事", "업무"},
	"家庭":  {"Family", "家族", "가족"},
	"个人":  {"Personal", "個人", "개인"},
	"情侣":  {"Couple", "カップル", "커플"},
	"大额":  {"Large", "高額", "고액"},
	"待还":  {"To Repay", "返済待ち", "상환 예정"},
	"可抵税": {"Tax Deductible", "控除対象", "세금 공제"},
}

// Preset 默认分类 / 标签在目标语言下的名字；缺译时保留中文
func Preset(locale, zh string) string {
	if locale == ZhCN || locale == "" {
		return zh
	}
	if e, ok := presetNames[zh]; ok {
		if s := pick(e, locale); s != "" {
			return s
		}
	}
	return zh
}
