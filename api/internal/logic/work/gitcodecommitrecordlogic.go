package work

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"apitools/api/internal/svc"
	"apitools/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GitCodeCommitRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGitCodeCommitRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GitCodeCommitRecordLogic {
	return &GitCodeCommitRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GitLab API 响应结构
type GitLabProject struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
}

type GitLabCommit struct {
	ID            string    `json:"id"`
	ShortID       string    `json:"short_id"`
	Title         string    `json:"title"`
	Message       string    `json:"message"`
	AuthorName    string    `json:"author_name"`
	AuthorEmail   string    `json:"author_email"`
	CommittedDate time.Time `json:"committed_date"`
	WebURL        string    `json:"web_url"`
}

type GitLabUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

func (l *GitCodeCommitRecordLogic) GitCodeCommitRecord(req *types.GitCommitRecordReq) (resp *types.GitCommitRecordResp, err error) {
	// 参数验证
	if err = l.validateRequest(req); err != nil {
		return &types.GitCommitRecordResp{
			Code:    400,
			Message: fmt.Sprintf("参数验证失败: %v", err),
		}, nil
	}

	// 获取配置
	gitlabUrl := req.GitlabUrl
	if gitlabUrl == "" {
		gitlabUrl = l.svcCtx.Config.GitLab.DefaultUrl
	}

	accessToken := req.AccessToken
	if accessToken == "" {
		accessToken = l.svcCtx.Config.GitLab.DefaultAccessToken
	}

	// 解析时间范围 - 支持 today 关键字和空值处理
	startDate, endDate, dateRange, err := l.parseDateRange(req.StartDate, req.EndDate)
	if err != nil {
		return &types.GitCommitRecordResp{
			Code:    400,
			Message: fmt.Sprintf("时间参数解析失败: %v", err),
		}, nil
	}

	// 创建 HTTP 客户端
	client := l.createHTTPClient()

	// 获取用户ID
	userID, err := l.getUserID(client, gitlabUrl, accessToken, req.Username)
	if err != nil {
		return &types.GitCommitRecordResp{
			Code:    500,
			Message: fmt.Sprintf("获取用户信息失败: %v", err),
		}, nil
	}

	l.Infof("获取到用户ID: %d (用户名: %s)", userID, req.Username)

	// 查询所有项目的提交记录
	var projectCommits []types.ProjectCommits
	totalCommits := int64(0)
	projectsWithCommits := int64(0)

	for _, projectPath := range req.Projects {
		l.Infof("正在查询项目: %s", projectPath)

		project, commits, err := l.getProjectCommits(client, gitlabUrl, accessToken, projectPath, req.Username, startDate, endDate)
		if err != nil {
			l.Errorf("查询项目 %s 失败: %v", projectPath, err)
			continue
		}

		if len(commits) > 0 {
			projectsWithCommits++
			totalCommits += int64(len(commits))

			projectCommit := types.ProjectCommits{
				ProjectId:   int64(project.ID),
				ProjectName: project.Name,
				ProjectPath: project.PathWithNamespace,
				ProjectUrl:  project.WebURL,
				Commits:     commits,
				CommitCount: int64(len(commits)),
			}
			projectCommits = append(projectCommits, projectCommit)

			l.Infof("项目 %s: 找到 %d 个提交", project.Name, len(commits))
		}
	}

	// 构建响应
	resp = &types.GitCommitRecordResp{
		Code:           200,
		Message:        "查询成功",
		Username:       req.Username,
		DateRange:      dateRange,
		ProjectCommits: projectCommits,
		Summary: types.GitCommitSummary{
			ProjectsWithCommits: projectsWithCommits,
			TotalCommits:        totalCommits,
			TotalProjects:       int64(len(req.Projects)),
			GitlabServer:        gitlabUrl,
		},
	}

	l.Infof("查询完成 - 总项目数: %d, 有提交的项目数: %d, 总提交数: %d",
		len(req.Projects), projectsWithCommits, totalCommits)

	return resp, nil
}

// validateRequest 验证请求参数
func (l *GitCodeCommitRecordLogic) validateRequest(req *types.GitCommitRecordReq) error {
	if len(req.Projects) == 0 {
		return fmt.Errorf("项目列表不能为空")
	}
	if req.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}
	return nil
}

// parseDateRange 解析时间范围
func (l *GitCodeCommitRecordLogic) parseDateRange(startDate, endDate string) (time.Time, time.Time, string, error) {
	var start, end time.Time
	var err error

	// 处理开始时间 - 支持 today 关键字和空值
	if startDate == "" || strings.ToLower(strings.TrimSpace(startDate)) == "today" {
		start = time.Now().Truncate(24 * time.Hour)
		l.Infof("使用今天作为开始时间: %s", start.Format("2006-01-02"))
	} else {
		start, err = time.Parse("2006-01-02", strings.TrimSpace(startDate))
		if err != nil {
			return start, end, "", fmt.Errorf("开始时间格式错误，请使用 YYYY-MM-DD 格式或 'today'，当前值: %s", startDate)
		}
		l.Infof("解析开始时间: %s", start.Format("2006-01-02"))
	}

	// 处理结束时间 - 支持 today 关键字和空值
	if endDate == "" || strings.ToLower(strings.TrimSpace(endDate)) == "today" {
		if strings.ToLower(strings.TrimSpace(endDate)) == "today" {
			// 如果结束时间明确指定为 "today"，使用今天
			end = time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour).Add(-time.Second)
			l.Infof("使用今天作为结束时间")
		} else {
			// 如果结束时间为空，使用开始时间的当天结束
			end = start.Add(24 * time.Hour).Add(-time.Second)
			l.Infof("结束时间为空，使用开始时间当天结束")
		}
	} else {
		end, err = time.Parse("2006-01-02", strings.TrimSpace(endDate))
		if err != nil {
			return start, end, "", fmt.Errorf("结束时间格式错误，请使用 YYYY-MM-DD 格式或 'today'，当前值: %s", endDate)
		}
		end = end.Add(24 * time.Hour).Add(-time.Second) // 当天结束
		l.Infof("解析结束时间: %s", end.Format("2006-01-02"))
	}

	// 验证时间范围
	if end.Before(start) {
		return start, end, "", fmt.Errorf("结束时间不能早于开始时间")
	}

	// 构建日期范围描述
	dateRange := start.Format("2006-01-02")
	if !end.Truncate(24 * time.Hour).Equal(start.Truncate(24 * time.Hour)) {
		dateRange += " 至 " + end.Truncate(24*time.Hour).Format("2006-01-02")
	}

	l.Infof("最终时间范围: %s", dateRange)
	return start, end, dateRange, nil
}

// createHTTPClient 创建 HTTP 客户端
func (l *GitCodeCommitRecordLogic) createHTTPClient() *http.Client {
	timeout := time.Duration(l.svcCtx.Config.GitLab.TimeoutSeconds) * time.Second
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

// getUserID 获取用户ID
func (l *GitCodeCommitRecordLogic) getUserID(client *http.Client, gitlabUrl, accessToken, username string) (int, error) {
	apiUrl := fmt.Sprintf("%s/api/v4/users", strings.TrimSuffix(gitlabUrl, "/"))

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("username", username)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("GitLab API 返回错误状态码: %d", resp.StatusCode)
	}

	var users []GitLabUser
	if err = json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return 0, err
	}

	if len(users) == 0 {
		return 0, fmt.Errorf("未找到用户: %s", username)
	}

	return users[0].ID, nil
}

// getProjectCommits 获取项目提交记录
func (l *GitCodeCommitRecordLogic) getProjectCommits(client *http.Client, gitlabUrl, accessToken, projectPath, username string, startDate, endDate time.Time) (*GitLabProject, []types.CommitInfo, error) {
	// 获取项目信息
	project, err := l.getProject(client, gitlabUrl, accessToken, projectPath)
	if err != nil {
		return nil, nil, err
	}

	// 获取提交记录
	commits, err := l.getCommits(client, gitlabUrl, accessToken, project.ID, username, startDate, endDate)
	if err != nil {
		return project, nil, err
	}

	// 转换为响应格式
	var commitInfos []types.CommitInfo
	for _, commit := range commits {
		commitInfo := types.CommitInfo{
			CommitId:      commit.ID,
			ShortId:       commit.ShortID,
			Title:         commit.Title,
			Message:       commit.Message,
			AuthorName:    commit.AuthorName,
			AuthorEmail:   commit.AuthorEmail,
			CommittedDate: commit.CommittedDate.Format("2006-01-02 15:04:05"),
			WebUrl:        fmt.Sprintf("%s/%s/-/commit/%s", strings.TrimSuffix(gitlabUrl, "/"), project.PathWithNamespace, commit.ID),
		}
		commitInfos = append(commitInfos, commitInfo)
	}

	return project, commitInfos, nil
}

// getProject 获取项目信息
func (l *GitCodeCommitRecordLogic) getProject(client *http.Client, gitlabUrl, accessToken, projectPath string) (*GitLabProject, error) {
	encodedPath := url.QueryEscape(projectPath)
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%s", strings.TrimSuffix(gitlabUrl, "/"), encodedPath)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("项目不存在或无访问权限: %s", projectPath)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("获取项目信息失败，状态码: %d", resp.StatusCode)
	}

	var project GitLabProject
	if err = json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, err
	}

	return &project, nil
}

// getCommits 获取提交记录
func (l *GitCodeCommitRecordLogic) getCommits(client *http.Client, gitlabUrl, accessToken string, projectID int, username string, startDate, endDate time.Time) ([]GitLabCommit, error) {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%d/repository/commits", strings.TrimSuffix(gitlabUrl, "/"), projectID)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("since", startDate.Format(time.RFC3339))
	q.Add("until", endDate.Format(time.RFC3339))
	q.Add("per_page", "100")
	q.Add("all", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("获取提交记录失败，状态码: %d", resp.StatusCode)
	}

	var allCommits []GitLabCommit
	if err = json.NewDecoder(resp.Body).Decode(&allCommits); err != nil {
		return nil, err
	}

	// 过滤指定用户的提交
	var userCommits []GitLabCommit
	for _, commit := range allCommits {
		if l.isUserCommit(commit, username) {
			userCommits = append(userCommits, commit)
		}
	}

	return userCommits, nil
}

// isUserCommit 检查是否为指定用户的提交
func (l *GitCodeCommitRecordLogic) isUserCommit(commit GitLabCommit, username string) bool {
	username = strings.ToLower(username)

	return strings.Contains(strings.ToLower(commit.AuthorEmail), username) ||
		strings.Contains(strings.ToLower(commit.AuthorName), username)
}
