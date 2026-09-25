package service

import (
	"errors"
	"time"

	"codelearn/model"
	"codelearn/repository"
	"codelearn/sandbox"
)

// ProjectService 项目实战工坊（U9）：多文件项目编辑/运行/完成
type ProjectService struct {
	repo *repository.Repository
}

func NewProjectService(repo *repository.Repository) *ProjectService {
	return &ProjectService{repo: repo}
}

var ErrProjectNotFound = errors.New("项目不存在")

// projectTemplates 各语言默认项目模板
var projectTemplates = map[string][]struct {
	Name    string
	Content string
}{
	"python": {
		{Name: "main.py", Content: `# 项目主文件
# 在这里编写你的项目代码

def main():
    print("Hello, CodeLearn Project!")

if __name__ == "__main__":
    main()
`},
	},
	"go": {
		{Name: "main.go", Content: `package main

import "fmt"

func main() {
	fmt.Println("Hello, CodeLearn Project!")
}
`},
		{Name: "go.mod", Content: `module codelearnproject

go 1.25
`},
	},
}

// ProjectTemplate 项目创建模板（含示例文件）
type ProjectTemplate struct {
	Language string
	Title    string
	Files    []struct {
		Name    string
		Content string
	}
}

// ListTemplates 返回可创建的项目模板列表
func (s *ProjectService) ListTemplates() []ProjectTemplate {
	out := []ProjectTemplate{
		{Language: "python", Title: "Python 小项目"},
		{Language: "go", Title: "Go 小项目"},
	}
	for i := range out {
		for _, f := range projectTemplates[out[i].Language] {
			out[i].Files = append(out[i].Files, struct {
				Name    string
				Content string
			}{Name: f.Name, Content: f.Content})
		}
	}
	return out
}

// Create 创建项目（course 为可选关联课程）
func (s *ProjectService) Create(userID, courseID uint, title, description, language, mainFile string) (*model.Project, error) {
	if language == "" {
		language = "python"
	}
	if mainFile == "" {
		if language == "go" {
			mainFile = "main.go"
		} else {
			mainFile = "main.py"
		}
	}
	courseTitle := ""
	if courseID > 0 {
		if c, err := s.repo.GetCourse(courseID); err == nil {
			courseTitle = c.Title
			if language == "" || language == "python" {
				language = c.Language
				if language == "py" {
					language = "python"
				}
				if language == "go" {
					mainFile = "main.go"
				} else {
					mainFile = "main.py"
				}
			}
		}
	}
	if title == "" {
		title = "我的项目"
	}

	project := &model.Project{
		UserID:      userID,
		CourseID:    courseID,
		CourseTitle: courseTitle,
		Title:       title,
		Description: description,
		Language:    language,
		MainFile:    mainFile,
		Status:      "in_progress",
	}
	if err := s.repo.CreateProject(project); err != nil {
		return nil, err
	}

	files := projectTemplates[language]
	models := make([]model.ProjectFile, 0, len(files))
	for i, f := range files {
		models = append(models, model.ProjectFile{
			ProjectID: project.ID,
			Name:      f.Name,
			Content:   f.Content,
			Order:     i,
		})
	}
	if err := s.repo.ReplaceProjectFiles(project.ID, models); err != nil {
		return nil, err
	}
	return project, nil
}

// List 列出用户项目
func (s *ProjectService) List(userID uint) ([]model.Project, error) {
	return s.repo.ListProjectsByUser(userID)
}

// Get 获取项目详情（含文件）
func (s *ProjectService) Get(userID, projectID uint) (*model.Project, []model.ProjectFile, error) {
	p, err := s.repo.GetProject(userID, projectID)
	if err != nil {
		return nil, nil, ErrProjectNotFound
	}
	files, err := s.repo.ListProjectFiles(projectID)
	if err != nil {
		return nil, nil, err
	}
	return p, files, nil
}

// SaveFiles 保存项目文件
func (s *ProjectService) SaveFiles(userID, projectID uint, mainFile string, files []model.ProjectFile) error {
	p, err := s.repo.GetProject(userID, projectID)
	if err != nil {
		return ErrProjectNotFound
	}
	if mainFile != "" {
		p.MainFile = mainFile
		if err := s.repo.UpdateProject(p); err != nil {
			return err
		}
	}
	for i := range files {
		files[i].ProjectID = projectID
	}
	return s.repo.ReplaceProjectFiles(projectID, files)
}

// Run 运行项目（多文件沙箱）
func (s *ProjectService) Run(userID, projectID uint) (*sandbox.RunResult, error) {
	p, err := s.repo.GetProject(userID, projectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	files, err := s.repo.ListProjectFiles(projectID)
	if err != nil {
		return nil, err
	}
	inputs := make([]sandbox.ProjectFileInput, 0, len(files))
	for _, f := range files {
		inputs = append(inputs, sandbox.ProjectFileInput{Name: f.Name, Content: f.Content})
	}
	res := sandbox.RunProject(p.Language, p.MainFile, inputs)

	p.RunCount++
	p.LastRunOutput = res.Output
	p.LastRunError = res.Error
	_ = s.repo.UpdateProject(p)
	return res, nil
}

// Complete 标记项目完成
func (s *ProjectService) Complete(userID, projectID uint) (*model.Project, error) {
	p, err := s.repo.GetProject(userID, projectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	now := time.Now()
	p.Status = "completed"
	p.CompletedAt = &now
	if err := s.repo.UpdateProject(p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete 删除项目
func (s *ProjectService) Delete(userID, projectID uint) error {
	if _, err := s.repo.GetProject(userID, projectID); err != nil {
		return ErrProjectNotFound
	}
	return s.repo.DeleteProject(userID, projectID)
}
