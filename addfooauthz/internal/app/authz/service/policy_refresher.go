package service

import (
	"context"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"

	"github.com/cage1016/gokit-istio-security/internal/app/authz/engine"
	storageV1 "github.com/cage1016/gokit-istio-security/internal/app/authz/storage/v1"
)

var ErrMessageBoxFull = errors.New("Message box full")

type PolicyRefresher interface {
	Refresh(context.Context) error
	RefreshAsync() error
}

type policyRefresher struct {
	log                      log.Logger
	store                    storageV1.Storage
	engine                   engine.Writer
	refreshRequests          chan policyRefresherMessageRefresh
	antiEntropyTimerDuration time.Duration
	changeNotifier           storageV1.PolicyChangeNotifier
}

type policyRefresherMessageRefresh struct {
	ctx    context.Context
	status chan error
}

func (m *policyRefresherMessageRefresh) Respond(err error) {
	if m.status != nil {
		select {
		case m.status <- err:
		default:
		}
		close(m.status)
	}
}

func (m *policyRefresherMessageRefresh) Err() error {
	return <-m.status
}

func NewPolicyRefresher(ctx context.Context, log log.Logger, store storageV1.Storage,
	engine engine.Writer) (PolicyRefresher, error) {
	changeNotifier, err := store.GetPolicyChangeNotifier(ctx)
	if err != nil {
		return nil, err
	}
	refresher := &policyRefresher{
		log:                      log,
		store:                    store,
		engine:                   engine,
		refreshRequests:          make(chan policyRefresherMessageRefresh, 1),
		antiEntropyTimerDuration: 10 * time.Second,
		changeNotifier:           changeNotifier,
	}
	go refresher.run(ctx)
	return refresher, nil
}

func (refresher *policyRefresher) run(ctx context.Context) {
	var lastPolicyChangeID string
	antiEntropyTimer := time.NewTimer(refresher.antiEntropyTimerDuration)
RUNLOOP:
	for {
		select {
		case <-ctx.Done():
			refresher.log.Log("done", "Policy refresher exiting")
			//refresher.log.WithError(ctx.Err()).Info("Policy refresher exiting")
			break RUNLOOP
		case <-refresher.changeNotifier.C():
			//refresher.log.Info("Received policy change notification")
			refresher.log.Log("changeNotifier", "Received policy change notification")
			var err error
			lastPolicyChangeID, err = refresher.refresh(context.Background(), lastPolicyChangeID)
			if err != nil {
				refresher.log.Log("err", err)
				//refresher.log.WithError(err).Warn("Failed to refresh policies")
			}
			if !antiEntropyTimer.Stop() {
				<-antiEntropyTimer.C
			}
		case m := <-refresher.refreshRequests:
			refresher.log.Log("refreshRequests", "Received local policy refresh request")
			//refresher.log.Info("Received local policy refresh request")
			var err error
			lastPolicyChangeID, err = refresher.refresh(m.ctx, lastPolicyChangeID)
			m.Respond(err)
			if !antiEntropyTimer.Stop() {
				<-antiEntropyTimer.C
			}
		case <-antiEntropyTimer.C:
			var err error
			lastPolicyChangeID, err = refresher.refresh(ctx, lastPolicyChangeID)
			if err != nil {
				refresher.log.Log("err", err)
				//refresher.log.WithError(err).Warn("Anti-entropy refresh failed")
			}
		}

		antiEntropyTimer.Reset(refresher.antiEntropyTimerDuration)
	}
	//refresher.log.Info("Shutting down policy refresh loop")
	refresher.log.Log("done", "Shutting down policy refresh loop")
	close(refresher.refreshRequests)
}

func (refresher *policyRefresher) refresh(ctx context.Context, lastPolicyChangeID string) (string, error) {
	//return "", nil
	//curPolicyID, err := refresher.store.GetPolicyChangeID(ctx)
	//if err != nil {
	//	refresher.log.WithError(err).Warn("Failed to get current policy change ID")
	//	return lastPolicyChangeID, err
	//}
	//if curPolicyID != lastPolicyChangeID {
	//	refresher.log.WithFields(logrus.Fields{
	//		"lastPolicyChangeID": lastPolicyChangeID,
	//		"curPolicyID":        curPolicyID,
	//	}).Debug("Refreshing engine store")
	//
	//	if err := refresher.updateEngineStore(ctx); err != nil {
	//		refresher.log.WithError(err).Warn("Failed to refresh engine store")
	//		return lastPolicyChangeID, err
	//	}
	//}
	//return curPolicyID, nil

	if err := refresher.updateEngineStore(ctx); err != nil {
		//refresher.log.WithError(err).Warn("Failed to refresh engine store")
		return "", err
	}
	return "", nil
}

func (refresher *policyRefresher) Refresh(ctx context.Context) error {
	m := policyRefresherMessageRefresh{
		ctx:    ctx,
		status: make(chan error, 1),
	}
	refresher.refreshRequests <- m
	return m.Err()
}

func (refresher *policyRefresher) RefreshAsync() error {
	m := policyRefresherMessageRefresh{
		ctx: context.Background(),
	}
	select {
	case refresher.refreshRequests <- m:
	default:
		//refresher.log.Warn("Refresher message box full")
		refresher.log.Log("warning", "Refresher message box full")
		return ErrMessageBoxFull
	}

	return nil
}

// updates OPA engine store with policy
func (refresher *policyRefresher) updateEngineStore(ctx context.Context) error {

	policiesRoles := make(map[string]interface{})
	{
		var userRoles []*storageV1.PoliciesUserRole
		var err error
		if userRoles, err = refresher.store.GetAllUserWithRoles(ctx); err != nil {
			return err
		}
		for _, u := range userRoles {
			policiesRoles[u.OrganizationIDStoreIDUserID] = strings.Split(u.RoleNames, ",")
		}
	}

	policiesPermissions := make(map[string]interface{})
	{
		type tempStruct struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		}
		var rolePermissions []*storageV1.PoliciesRolePermission
		var err error
		if rolePermissions, err = refresher.store.GetAllRolesWithPermission(ctx); err != nil {
			return err
		}
		ms := make(map[string][]*tempStruct)
		for _, rolePermission := range rolePermissions {
			t := &tempStruct{Method: rolePermission.Action, Path: rolePermission.Resource}
			if _, ok := ms[rolePermission.RoleName]; !ok {
				ms[rolePermission.RoleName] = []*tempStruct{t}
			} else {
				ms[rolePermission.RoleName] = append(ms[rolePermission.RoleName], t)
			}
		}
		for k, v := range ms {
			policiesPermissions[k] = v
		}
	}

	input := map[string]interface{}{
		"userRoles":       policiesRoles,
		"rolePermissions": policiesPermissions,
	}

	return refresher.engine.SetUserRolesAndPermissions(ctx, input)
}

//func (refresher *policyRefresher) getPolicyMap(ctx context.Context) (map[string]interface{}, error) {
//	var policies []*storageV1.Policy
//	var err error
//
//	if policies, err = refresher.store.ListPolicies(ctx); err != nil {
//		return nil, err
//	}
//	refresher.log.Infof("initializing OPA store with %d V2 policies", len(policies))
//
//	policies = append(policies, SystemPolicies()...)
//
//	// OPA requires this format
//	data := make(map[string]interface{})
//	for _, p := range policies {
//
//		statements := make(map[string]interface{})
//		for _, st := range p.Statements {
//			stmt := map[string]interface{}{
//				"effect":   st.Effect.String(),
//				"projects": st.Projects,
//			}
//			// Only set these if provided
//			if st.Role != "" {
//				stmt["role"] = st.Role
//			}
//			if len(st.Actions) != 0 {
//				stmt["actions"] = st.Actions
//			}
//			if len(st.Resources) != 0 {
//				stmt["resources"] = st.Resources
//			}
//			statements[st.ID.String()] = stmt
//		}
//
//		members := make([]string, len(p.Members))
//		for i, member := range p.Members {
//			members[i] = member.Name
//		}
//
//		data[p.ID] = map[string]interface{}{
//			"type":       p.Type.String(),
//			"members":    members,
//			"statements": statements,
//		}
//	}
//	return data, nil
//}
//
//func (refresher *policyRefresher) getRoleMap(ctx context.Context) (map[string]interface{}, error) {
//	var roles []*storageV1.Role
//	var err error
//	if roles, err = refresher.store.ListRoles(ctx); err != nil {
//		return nil, err
//	}
//	refresher.log.Infof("initializing OPA store with %d V2 roles", len(roles))
//
//	// OPA requires this format
//	data := make(map[string]interface{})
//	for _, r := range roles {
//		data[r.ID] = map[string]interface{}{
//			"actions": r.Actions,
//		}
//	}
//	return data, nil
//}
//
//func (refresher *policyRefresher) getIAMVersion(ctx context.Context) (api.Version, error) {
//	var vsn api.Version
//	ms, err := refresher.store.MigrationStatus(ctx)
//	if err != nil {
//		return vsn, err
//	}
//	switch ms {
//	case storageV1.Successful:
//		vsn = api.Version{Major: api.Version_V2, Minor: api.Version_V0}
//	case storageV1.SuccessfulBeta1:
//		vsn = api.Version{Major: api.Version_V2, Minor: api.Version_V1}
//	default:
//		vsn = api.Version{Major: api.Version_V1, Minor: api.Version_V0}
//	}
//	return vsn, nil
//}
//
//func (refresher *policyRefresher) getRuleMap(ctx context.Context) (map[string][]storageV1.Rule, error) {
//	rules, err := refresher.store.ListRules(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	ruleMap := make(map[string][]storageV1.Rule)
//	for _, r := range rules {
//		if _, ok := ruleMap[r.ProjectID]; !ok {
//			ruleMap[r.ProjectID] = make([]storageV1.Rule, 0)
//		}
//
//		ruleMap[r.ProjectID] = append(ruleMap[r.ProjectID], *r)
//	}
//
//	return ruleMap, nil
//}
