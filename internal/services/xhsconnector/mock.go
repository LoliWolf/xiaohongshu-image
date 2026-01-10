package xhsconnector

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type MockConnector struct {
	mu       sync.RWMutex
	comments map[string][]Comment
}

func NewMockConnector() *MockConnector {
	m := &MockConnector{
		comments: make(map[string][]Comment),
	}
	m.initMockComments()
	return m
}

func (m *MockConnector) initMockComments() {
	now := time.Now()

	mockComments := []Comment{
		{
			CommentID:        "mock_001",
			UserName:         "测试用户1",
			Content:          "帮我画一张可爱的猫咪图片，邮箱：test1@example.com",
			CommentCreatedAt: now.Add(-2 * time.Hour),
		},
		{
			CommentID:        "mock_002",
			UserName:         "测试用户2",
			Content:          "能生成一个视频吗？主题是海边日落，contact@demo.com",
			CommentCreatedAt: now.Add(-1 * time.Hour),
		},
		{
			CommentID:        "mock_003",
			UserName:         "测试用户3",
			Content:          "这个笔记真好看！",
			CommentCreatedAt: now.Add(-30 * time.Minute),
		},
		{
			CommentID:        "mock_004",
			UserName:         "测试用户4",
			Content:          "AI生成一张赛博朋克风格的图片，myemail@company.com",
			CommentCreatedAt: now.Add(-15 * time.Minute),
		},
		{
			CommentID:        "mock_005",
			UserName:         "测试用户5",
			Content:          "做个视频，内容是城市夜景，sendto@user.org",
			CommentCreatedAt: now.Add(-5 * time.Minute),
		},
		{
			CommentID:        "mock_006",
			UserName:         "测试用户6",
			Content:          "出图！风景画，风格是油画，art@studio.com",
			CommentCreatedAt: now,
		},
	}

	m.comments["default"] = mockComments
}

func (m *MockConnector) ListComments(ctx context.Context, noteIDOrURL string, cursor string) (*ListCommentsResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	comments, exists := m.comments[noteIDOrURL]
	if !exists {
		comments = m.comments["default"]
	}

	var startIndex int
	if cursor != "" {
		for i, c := range comments {
			if c.CommentID == cursor {
				startIndex = i + 1
				break
			}
		}
	}

	endIndex := startIndex + 50
	if endIndex > len(comments) {
		endIndex = len(comments)
	}

	resultComments := comments[startIndex:endIndex]

	var nextCursor string
	var hasMore bool
	if endIndex < len(comments) {
		nextCursor = comments[endIndex-1].CommentID
		hasMore = true
	}

	time.Sleep(100*time.Millisecond + time.Duration(rand.Intn(200))*time.Millisecond)

	return &ListCommentsResult{
		Comments:   resultComments,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (m *MockConnector) AddComment(noteIDOrURL string, comment Comment) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.comments[noteIDOrURL]; !exists {
		m.comments[noteIDOrURL] = []Comment{}
	}

	m.comments[noteIDOrURL] = append(m.comments[noteIDOrURL], comment)
}

func (m *MockConnector) GenerateNewComment(noteIDOrURL string) Comment {
	prompts := []string{
		"帮我生成一张美食图片，邮箱：",
		"AI做个视频，主题是旅行，联系邮箱：",
		"出图！风格是动漫，邮箱：",
		"生成一个视频，内容是宠物，邮箱：",
		"帮我画一张风景画，邮箱：",
	}

	emails := []string{
		"user1@test.com",
		"demo@example.org",
		"contact@company.com",
		"test@demo.net",
		"hello@world.com",
	}

	rand.Seed(time.Now().UnixNano())

	prompt := prompts[rand.Intn(len(prompts))]
	email := emails[rand.Intn(len(emails))]

	return Comment{
		CommentID:        fmt.Sprintf("mock_%d", time.Now().UnixNano()),
		UserName:         fmt.Sprintf("随机用户%d", rand.Intn(1000)),
		Content:          prompt + email,
		CommentCreatedAt: time.Now(),
	}
}
