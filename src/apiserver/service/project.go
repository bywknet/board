package service

import (
	"errors"
	"git/inspursoft/board/src/common/dao"
	"git/inspursoft/board/src/common/model"
)

func CreateProject(project model.Project) (bool, error) {
	projectID, err := dao.AddProject(project)
	if err != nil {
		return false, err
	}

	projectMember := model.ProjectMember{
		ProjectID: projectID,
		UserID:    int64(project.OwnerID),
		RoleID:    model.ProjectAdmin,
	}
	projectMemberID, err := dao.InsertOrUpdateProjectMember(projectMember)
	if err != nil {
		return false, errors.New("failed to create project member")
	}
	return (projectID != 0 && projectMemberID != 0), nil
}

func GetProject(project model.Project, selectedFields ...string) (*model.Project, error) {
	p, err := dao.GetProject(project, selectedFields...)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func ProjectExists(projectName string) (bool, error) {
	query := model.Project{Name: projectName}
	project, err := dao.GetProject(query, "name")
	if err != nil {
		return false, err
	}
	return (project != nil && project.ID != 0), nil
}

func ProjectExistsByID(projectID int64) (bool, error) {
	query := model.Project{ID: projectID, Deleted: 0}
	project, err := dao.GetProject(query, "id", "deleted")
	if err != nil {
		return false, err
	}
	return (project != nil && project.Name != ""), nil
}

func UpdateProject(project model.Project, fieldNames ...string) (bool, error) {
	if project.ID == 0 {
		return false, errors.New("no Project ID provided")
	}
	_, err := dao.UpdateProject(project, fieldNames...)
	if err != nil {
		return false, err
	}
	return true, nil
}

func GetAllProjects(query model.Project) ([]*model.Project, error) {
	return dao.GetAllProjects(query)
}

func GetProjectsByUser(query model.Project, userID int64) ([]*model.Project, error) {
	return dao.GetProjectsByUser(query, userID)
}

func DeleteProject(projectID int64) (bool, error) {
	project := model.Project{ID: projectID, Deleted: 1}
	_, err := dao.UpdateProject(project, "deleted")
	if err != nil {
		return false, err
	}
	return true, nil
}
