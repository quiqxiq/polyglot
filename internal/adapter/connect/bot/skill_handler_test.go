package bot_test

import (
	"context"
	"os"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	botConnect "github.com/quixiq/polyglot/internal/adapter/connect/bot"
	"github.com/quixiq/polyglot/internal/adapter/storage"
	"github.com/quixiq/polyglot/internal/domain/skill"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
)

type mockSkillMetadataRepo struct {
	records map[string]*skill.SkillMetadataRecord
}

func (m *mockSkillMetadataRepo) ListSkills(ctx context.Context, userID string) ([]skill.SkillMetadataRecord, error) {
	var list []skill.SkillMetadataRecord
	for _, r := range m.records {
		list = append(list, *r)
	}
	return list, nil
}

func (m *mockSkillMetadataRepo) GetSkill(ctx context.Context, userID, name string) (*skill.SkillMetadataRecord, error) {
	if r, ok := m.records[name]; ok {
		return r, nil
	}
	return nil, skill.ErrSkillNotFound
}

func (m *mockSkillMetadataRepo) SaveSkillMetadata(ctx context.Context, rec *skill.SkillMetadataRecord) error {
	m.records[rec.Name] = rec
	return nil
}

func (m *mockSkillMetadataRepo) DeleteSkillMetadata(ctx context.Context, userID, name string) error {
	delete(m.records, name)
	return nil
}

func (m *mockSkillMetadataRepo) ListGitSkills(ctx context.Context) ([]skill.SkillMetadataRecord, error) {
	return nil, nil
}

func (m *mockSkillMetadataRepo) ToggleSkillEnabled(ctx context.Context, userID, name string, enabled bool) error {
	if r, ok := m.records[name]; ok {
		r.Enabled = enabled
		return nil
	}
	return skill.ErrSkillNotFound
}

func (m *mockSkillMetadataRepo) GetGlobalSystemPrompt(ctx context.Context) (string, error) {
	return "Global prompt", nil
}

func (m *mockSkillMetadataRepo) SaveGlobalSystemPrompt(ctx context.Context, content string) error {
	return nil
}

func TestSkillConnectHandler_E2E(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skill_connect_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	fsStore, err := storage.NewFSSkillStore(tempDir)
	require.NoError(t, err)

	repo := &mockSkillMetadataRepo{records: make(map[string]*skill.SkillMetadataRecord)}
	uc := skillUC.NewManageSkillUseCase(repo, fsStore, nil)
	handler := botConnect.NewSkillConnectHandler(uc)
	ctx := context.Background()

	// 1. Create skill
	createResp, err := handler.CreateSkill(ctx, connect.NewRequest(&devicepb.CreateSkillRequest{
		UserId:        "user-1",
		Name:          "netops-diag",
		Description:   "Diagnosa jaringan mikrotik",
		Content:       "# Langkah Diagnosa\n/ping 8.8.8.8",
		License:       "MIT",
		Compatibility: ">=1.0.0",
		AllowedTools:  "mcp_device",
	}))
	require.NoError(t, err)
	assert.Equal(t, "netops-diag", createResp.Msg.Skill.Name)

	// 2. List skills
	listResp, err := handler.ListSkills(ctx, connect.NewRequest(&devicepb.ListSkillsRequest{
		UserId: "user-1",
	}))
	require.NoError(t, err)
	assert.Len(t, listResp.Msg.Skills, 1)

	// 3. Save Resource
	saveResResp, err := handler.SaveResource(ctx, connect.NewRequest(&devicepb.SaveResourceRequest{
		SkillId: "netops-diag",
		Path:    "scripts/ping.rsc",
		Data:    []byte("/tool ping 1.1.1.1 count=5"),
	}))
	require.NoError(t, err)
	assert.True(t, saveResResp.Msg.Success)

	// 4. List Resources
	listResResp, err := handler.ListResources(ctx, connect.NewRequest(&devicepb.ListResourcesRequest{
		SkillId: "netops-diag",
	}))
	require.NoError(t, err)
	assert.Len(t, listResResp.Msg.Resources, 1)

	// 5. Export Skill
	exportResp, err := handler.ExportSkill(ctx, connect.NewRequest(&devicepb.ExportSkillRequest{
		Id: "netops-diag",
	}))
	require.NoError(t, err)
	assert.NotEmpty(t, exportResp.Msg.Archive)

	// 6. Delete Skill
	delResp, err := handler.DeleteSkill(ctx, connect.NewRequest(&devicepb.DeleteSkillRequest{
		UserId: "user-1",
		Id:     "netops-diag",
	}))
	require.NoError(t, err)
	assert.True(t, delResp.Msg.Success)
}
