package database

import (
	"time"

	"only4.fun/backend/internal/id"

	"github.com/jmoiron/sqlx"
)

type DeploymentStore struct {
	db *sqlx.DB
}

func NewDeploymentStore(db *Database) *DeploymentStore {
	return &DeploymentStore{db: db.DB()}
}

func (s *DeploymentStore) Create(projectID string, status string, log string) (*Deployment, error) {
	newID := id.New()
	if _, err := s.db.Exec(`INSERT INTO deployments (id, project_id, status, log) VALUES (?, ?, ?, ?)`, newID, projectID, status, log); err != nil {
		return nil, err
	}
	dep := Deployment{
		ID:        newID,
		ProjectID: projectID,
		Status:    status,
		Log:       log,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return &dep, nil
}

func (s *DeploymentStore) GetByProjectID(projectID string) ([]*Deployment, error) {
	var deployments []*Deployment
	query := `SELECT * FROM deployments WHERE project_id = ? ORDER BY created_at DESC`
	err := s.db.Select(&deployments, query, projectID)
	return deployments, err
}

func (s *DeploymentStore) GetLatest(projectID string) (*Deployment, error) {
	var deployment Deployment
	query := `SELECT * FROM deployments WHERE project_id = ? ORDER BY created_at DESC LIMIT 1`
	err := s.db.Get(&deployment, query, projectID)
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (s *DeploymentStore) GetByID(deploymentID string) (*Deployment, error) {
	var deployment Deployment
	query := `SELECT * FROM deployments WHERE id = ?`
	if err := s.db.Get(&deployment, query, deploymentID); err != nil {
		return nil, err
	}
	return &deployment, nil
}

// SetRelease mencatat folder rilis yang dipublikasikan oleh satu deployment.
func (s *DeploymentStore) SetRelease(deploymentID, release string) error {
	_, err := s.db.Exec(`UPDATE deployments SET release = ? WHERE id = ?`, release, deploymentID)
	return err
}

func (s *DeploymentStore) ClearByProjectID(projectID string) error {
	_, err := s.db.Exec(`DELETE FROM deployments WHERE project_id = ?`, projectID)
	return err
}

func (s *DeploymentStore) UpdateStatus(deploymentID string, status string, log string) error {
	query := `UPDATE deployments SET status = ?, log = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, status, log, time.Now(), deploymentID)
	return err
}
