package ualert

import (
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/util"

	"github.com/grafana/grafana/pkg/infra/metrics"
	"github.com/grafana/grafana/pkg/models"
)

// getOrCreateGeneralFolder returns the general folder under the specific organisation
// If the general folder does not exist it creates it.
func (m *migration) getOrCreateGeneralFolder(orgID int64) (*dashboard, error) {
	// there is a unique constraint on org_id, folder_id, title
	// there are no nested folders so the parent folder id is always 0
	dashboard := dashboard{OrgId: orgID, FolderId: 0, Title: GENERAL_FOLDER}
	has, err := m.sess.Get(&dashboard)
	if err != nil {
		return nil, err
	} else if !has {
		// create folder
		result, err := m.createFolder(orgID, GENERAL_FOLDER)
		if err != nil {
			return nil, err
		}

		return result, nil
	}
	return &dashboard, nil
}

// based on sqlstore.saveDashboard()
// it should be called from inside a transaction
func (m *migration) createFolder(orgID int64, title string) (*dashboard, error) {
	cmd := saveFolderCommand{
		OrgId:    orgID,
		FolderId: 0,
		IsFolder: true,
		Dashboard: simplejson.NewFromAny(map[string]interface{}{
			"title": title,
		}),
	}
	dash := cmd.getDashboardModel()

	uid, err := m.generateNewDashboardUid(dash.OrgId)
	if err != nil {
		return nil, err
	}
	dash.setUid(uid)

	parentVersion := dash.Version
	dash.setVersion(1)
	dash.Created = time.Now()
	dash.CreatedBy = FOLDER_CREATED_BY
	dash.Updated = time.Now()
	dash.UpdatedBy = FOLDER_CREATED_BY
	metrics.MApiDashboardInsert.Inc()

	if _, err = m.sess.Insert(dash); err != nil {
		return nil, err
	}

	dashVersion := &models.DashboardVersion{
		DashboardId:   dash.Id,
		ParentVersion: parentVersion,
		RestoredFrom:  cmd.RestoredFrom,
		Version:       dash.Version,
		Created:       time.Now(),
		CreatedBy:     dash.UpdatedBy,
		Message:       cmd.Message,
		Data:          dash.Data,
	}

	// insert version entry
	if _, err := m.sess.Insert(dashVersion); err != nil {
		return nil, err
	}
	return dash, nil
}

func (m *migration) generateNewDashboardUid(orgId int64) (string, error) {
	for i := 0; i < 3; i++ {
		uid := util.GenerateShortUID()

		exists, err := m.sess.Where("org_id=? AND uid=?", orgId, uid).Get(&models.Dashboard{})
		if err != nil {
			return "", err
		}

		if !exists {
			return uid, nil
		}
	}

	return "", models.ErrDashboardFailedGenerateUniqueUid
}

// based on SQLStore.UpdateDashboardACL()
// it should be called from inside a transaction
func (m *migration) setACL(orgID int64, dashboardID int64, items []*models.DashboardAcl) error {
	if dashboardID <= 0 {
		return fmt.Errorf("folder id must be greater than zero for a folder permission")
	}
	for _, item := range items {
		if item.UserID == 0 && item.TeamID == 0 && (item.Role == nil || !item.Role.IsValid()) {
			return models.ErrDashboardAclInfoMissing
		}

		item.OrgID = orgID
		item.DashboardID = dashboardID
		item.Created = time.Now()
		item.Updated = time.Now()

		m.sess.Nullable("user_id", "team_id")
		if _, err := m.sess.Insert(item); err != nil {
			return err
		}
	}

	// Update dashboard HasAcl flag
	dashboard := models.Dashboard{HasAcl: true}
	_, err := m.sess.Cols("has_acl").Where("id=?", dashboardID).Update(&dashboard)
	return err
}

// based on SQLStore.GetDashboardAclInfoList()
func (m *migration) getACL(orgID, dashboardID int64) ([]*models.DashboardAcl, error) {
	var err error

	falseStr := m.mg.Dialect.BooleanStr(false)

	result := make([]*models.DashboardAcl, 0)
	rawSQL := `
			-- get distinct permissions for the dashboard and its parent folder
			SELECT DISTINCT
				da.user_id,
				da.team_id,
				da.permission,
				da.role
			FROM dashboard as d
				LEFT JOIN dashboard folder on folder.id = d.folder_id
				LEFT JOIN dashboard_acl AS da ON
				da.dashboard_id = d.id OR
				da.dashboard_id = d.folder_id  OR
				(
					-- include default permissions --
					da.org_id = -1 AND (
					  (folder.id IS NOT NULL AND folder.has_acl = ` + falseStr + `) OR
					  (folder.id IS NULL AND d.has_acl = ` + falseStr + `)
					)
				)
			WHERE d.org_id = ? AND d.id = ? AND da.id IS NOT NULL
			ORDER BY da.id ASC
			`
	err = m.sess.SQL(rawSQL, orgID, dashboardID).Find(&result)
	return result, err
}
