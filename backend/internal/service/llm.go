package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybookkeeping/internal/config"
	"mybookkeeping/internal/model"
)

type LLMService struct {
	cfg *config.Config
	db  *gorm.DB
}

func NewLLMService(cfg *config.Config, db *gorm.DB) *LLMService {
	return &LLMService{cfg: cfg, db: db}
}

func (s *LLMService) Enabled() bool {
	return s.cfg.DeepSeekAPIKey != ""
}

type RecognizedTx struct {
	Type          string   `json:"type"`
	AmountYuan    float64  `json:"amountYuan"`
	AmountFen     int64    `json:"amountFen"`
	AccountName   string   `json:"accountName"`
	AccountID     *uint64  `json:"accountId"`
	ToAccountName string   `json:"toAccountName,omitempty"`
	ToAccountID   *uint64  `json:"toAccountId,omitempty"`
	CategoryName  string   `json:"categoryName"`
	CategoryID    *uint64  `json:"categoryId"`
	Remark        string   `json:"remark"`
	TagNames      []string `json:"tagNames"`
	TagIDs        []uint64 `json:"tagIds"`
	HappenedAt    string   `json:"happenedAt,omitempty"`
	Raw           string   `json:"raw,omitempty"`
}

type llmChatReq struct {
	Model          string       `json:"model"`
	Messages       []llmMessage `json:"messages"`
	Temperature    float64      `json:"temperature"`
	ResponseFormat *llmRespFmt  `json:"response_format,omitempty"`
}

// Content 可以是 string（纯文本）或 []map（多模态）
type llmMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type llmRespFmt struct {
	Type string `json:"type"`
}

type llmChatResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type llmJSON struct {
	Type       string   `json:"type"`
	Amount     float64  `json:"amount"`
	Account    string   `json:"account"`
	ToAccount  string   `json:"toAccount"`
	Category   string   `json:"category"`
	Remark     string   `json:"remark"`
	Tags       []string `json:"tags"`
	HappenedAt string   `json:"happenedAt"`
}

type userCatalog struct {
	system      string
	batchSystem string
	accMap      map[string]uint64
	catMap      map[string]uint64
	tagMap      map[string]uint64
}

func (s *LLMService) buildCatalog(userID uint64) (*userCatalog, error) {
	var accounts []model.Account
	if err := s.db.Where("user_id = ? AND archived = ?", userID, false).Order("sort asc").Find(&accounts).Error; err != nil {
		return nil, err
	}
	var cats []model.Category
	if err := s.db.Where("user_id = ?", userID).Order("sort asc").Find(&cats).Error; err != nil {
		return nil, err
	}
	var tags []model.Tag
	if err := s.db.Where("user_id = ?", userID).Order("sort asc").Find(&tags).Error; err != nil {
		return nil, err
	}

	accNames := make([]string, 0, len(accounts))
	accMap := map[string]uint64{}
	for _, a := range accounts {
		accNames = append(accNames, a.Name)
		accMap[a.Name] = a.ID
	}
	expNames, incNames := []string{}, []string{}
	catMap := map[string]uint64{}
	for _, c := range cats {
		catMap[c.Name] = c.ID
		if c.Kind == "expense" {
			expNames = append(expNames, c.Name)
		} else {
			incNames = append(incNames, c.Name)
		}
	}
	tagNames := make([]string, 0, len(tags))
	tagMap := map[string]uint64{}
	for _, t := range tags {
		tagNames = append(tagNames, t.Name)
		tagMap[t.Name] = t.ID
	}

	system := fmt.Sprintf(`你是记账助手。根据输入（文字描述或票据/小票/支付截图）提取一笔交易，只输出 JSON，不要其它文字。
字段：
- type: expense|income|transfer
- amount: 数字（元，可小数；票据取应付/实付总额）
- account: 账户名（尽量从列表选；截图里有支付宝/微信/银行卡可对应）
- toAccount: 转账时的转入账户，否则空字符串
- category: 分类名（尽量从列表选；转账可空）
- remark: 备注（可含商户名）
- tags: 标签名数组（尽量从列表选）
- happenedAt: RFC3339 或空（图中有日期则填）

当前时间：%s
支出分类：%s
收入分类：%s
账户：%s
标签：%s`,
		time.Now().Format(time.RFC3339),
		strings.Join(expNames, "、"),
		strings.Join(incNames, "、"),
		strings.Join(accNames, "、"),
		strings.Join(tagNames, "、"),
	)

	batchSystem := fmt.Sprintf(`你是记账助手。用户可能一次描述多笔收支/转账。请拆成多笔交易，只输出 JSON，不要其它文字。
输出格式：
{"transactions":[{...},{...}]}
每笔字段：
- type: expense|income|transfer
- amount: 数字（元，可小数）
- account: 账户名（尽量从下列列表选）
- toAccount: 转账时的转入账户，否则空字符串；信用卡还款视为 transfer（银行卡→信用卡），不要记成 expense
- category: 分类名（尽量从列表选；转账可空）
- remark: 备注
- tags: 标签名数组
- happenedAt: RFC3339 或空

规则：
1. 每笔独立交易单独一项；无法识别则返回 {"transactions":[]}
2. 不要把多笔合并成一笔
3. 金额必须为正数

当前时间：%s
支出分类：%s
收入分类：%s
账户：%s
标签：%s`,
		time.Now().Format(time.RFC3339),
		strings.Join(expNames, "、"),
		strings.Join(incNames, "、"),
		strings.Join(accNames, "、"),
		strings.Join(tagNames, "、"),
	)

	return &userCatalog{
		system: system, batchSystem: batchSystem,
		accMap: accMap, catMap: catMap, tagMap: tagMap,
	}, nil
}

func (s *LLMService) mapResult(content string, cat *userCatalog) (*RecognizedTx, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var parsed llmJSON
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("模型返回无法解析: %s", content)
	}

	out := &RecognizedTx{
		Type:          parsed.Type,
		AmountYuan:    parsed.Amount,
		AmountFen:     int64(parsed.Amount*100 + 0.5),
		AccountName:   parsed.Account,
		ToAccountName: parsed.ToAccount,
		CategoryName:  parsed.Category,
		Remark:        parsed.Remark,
		TagNames:      parsed.Tags,
		HappenedAt:    parsed.HappenedAt,
		Raw:           content,
	}
	if out.Type == "" {
		out.Type = "expense"
	}
	if id, ok := cat.accMap[parsed.Account]; ok {
		out.AccountID = &id
	}
	if parsed.ToAccount != "" {
		if id, ok := cat.accMap[parsed.ToAccount]; ok {
			out.ToAccountID = &id
		}
	}
	if id, ok := cat.catMap[parsed.Category]; ok {
		out.CategoryID = &id
	}
	for _, name := range parsed.Tags {
		if id, ok := cat.tagMap[name]; ok {
			out.TagIDs = append(out.TagIDs, id)
		}
	}
	return out, nil
}

func (s *LLMService) RecognizeText(ctx context.Context, userID uint64, text string) (*RecognizedTx, error) {
	if !s.Enabled() {
		return nil, errors.New("未配置 DEEPSEEK_API_KEY，AI 识别不可用")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("请输入要识别的文字")
	}
	cat, err := s.buildCatalog(userID)
	if err != nil {
		return nil, err
	}
	content, err := s.chat(ctx, s.cfg.DeepSeekModel, cat.system, text)
	if err != nil {
		return nil, err
	}
	return s.mapResult(content, cat)
}

// RecognizeTextBatch 从一段文字中拆出多笔交易。
func (s *LLMService) RecognizeTextBatch(ctx context.Context, userID uint64, text string) ([]*RecognizedTx, error) {
	if !s.Enabled() {
		return nil, errors.New("未配置 DEEPSEEK_API_KEY，AI 识别不可用")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("请输入要识别的文字")
	}
	cat, err := s.buildCatalog(userID)
	if err != nil {
		return nil, err
	}
	content, err := s.chat(ctx, s.cfg.DeepSeekModel, cat.batchSystem, text)
	if err != nil {
		return nil, err
	}
	return s.mapBatchResult(content, cat)
}

func (s *LLMService) mapBatchResult(content string, cat *userCatalog) ([]*RecognizedTx, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var wrapped struct {
		Transactions []llmJSON `json:"transactions"`
	}
	if err := json.Unmarshal([]byte(content), &wrapped); err != nil {
		// 兼容直接返回数组
		var arr []llmJSON
		if err2 := json.Unmarshal([]byte(content), &arr); err2 != nil {
			return nil, fmt.Errorf("模型返回无法解析: %s", content)
		}
		wrapped.Transactions = arr
	}

	out := make([]*RecognizedTx, 0, len(wrapped.Transactions))
	for _, parsed := range wrapped.Transactions {
		raw, _ := json.Marshal(parsed)
		item, err := s.mapResult(string(raw), cat)
		if err != nil {
			continue
		}
		if item.AmountYuan <= 0 && item.AmountFen <= 0 {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

// RecognizeImage 使用多模态模型识别票据/截图。imageData 为原始图片字节。
func (s *LLMService) RecognizeImage(ctx context.Context, userID uint64, imageData []byte, mimeType, hint string) (*RecognizedTx, error) {
	if !s.Enabled() {
		return nil, errors.New("未配置 DEEPSEEK_API_KEY，AI 识别不可用")
	}
	if len(imageData) == 0 {
		return nil, errors.New("请上传图片")
	}
	if len(imageData) > 12<<20 {
		return nil, errors.New("图片过大（上限约 12MB）")
	}
	if mimeType == "" || mimeType == "application/octet-stream" {
		mimeType = detectImageMIME(imageData)
	}
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
	default:
		return nil, errors.New("仅支持 JPEG / PNG / GIF / WebP")
	}

	cat, err := s.buildCatalog(userID)
	if err != nil {
		return nil, err
	}

	prompt := "请识别这张票据/小票/支付截图，提取记账信息并输出 JSON。"
	if h := strings.TrimSpace(hint); h != "" {
		prompt += "\n补充说明：" + h
	}

	b64 := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, b64)
	userContent := []map[string]any{
		{"type": "text", "text": prompt},
		{"type": "image_url", "image_url": map[string]any{"url": dataURL, "detail": "high"}},
	}

	model := s.cfg.DeepSeekVisionModel
	if model == "" {
		model = "deepseek-flash"
	}
	content, err := s.chat(ctx, model, cat.system, userContent)
	if err != nil {
		return nil, err
	}
	return s.mapResult(content, cat)
}

func detectImageMIME(data []byte) string {
	if len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff {
		return "image/jpeg"
	}
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4e && data[3] == 0x47 {
		return "image/png"
	}
	if len(data) >= 6 && (string(data[0:6]) == "GIF87a" || string(data[0:6]) == "GIF89a") {
		return "image/gif"
	}
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "image/webp"
	}
	return "image/jpeg"
}

func (s *LLMService) chat(ctx context.Context, model, system string, userContent any) (string, error) {
	base := strings.TrimRight(s.cfg.DeepSeekBaseURL, "/")
	url := base + "/v1/chat/completions"
	body := llmChatReq{
		Model: model,
		Messages: []llmMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: userContent},
		},
		Temperature:    0.1,
		ResponseFormat: &llmRespFmt{Type: "json_object"},
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.DeepSeekAPIKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var parsed llmChatResp
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("LLM 响应异常: %s", string(data))
	}
	if parsed.Error != nil {
		return "", errors.New(parsed.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("LLM HTTP %d: %s", resp.StatusCode, string(data))
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("LLM 无返回内容")
	}
	return parsed.Choices[0].Message.Content, nil
}
