package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
)

func TestGroupService_CreateGroup(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub)

	creator := uuid.New()
	memberA := uuid.New()
	memberB := uuid.New()
	userRepo.addUser(&model.User{ID: creator, Phone: "+628111", Name: "Creator"})
	userRepo.addUser(&model.User{ID: memberA, Phone: "+628222", Name: "MemberA"})
	userRepo.addUser(&model.User{ID: memberB, Phone: "+628333", Name: "MemberB"})

	t.Run("success", func(t *testing.T) {
		chat, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
			Name:      "Tim Proyek",
			Icon:      "💼",
			MemberIDs: []uuid.UUID{memberA, memberB},
		})
		require.NoError(t, err)
		assert.Equal(t, model.ChatTypeGroup, chat.Type)
		assert.Equal(t, "Tim Proyek", chat.Name)
		assert.Equal(t, "💼", chat.Icon)

		// Should have 3 members total
		members, _ := chatRepo.GetMembers(context.Background(), chat.ID)
		assert.Len(t, members, 3)

		// System message should have been created
		msgs := msgRepo.byChat[chat.ID]
		assert.GreaterOrEqual(t, len(msgs), 1)
		assert.Equal(t, model.MessageTypeSystem, msgs[0].Type)
	})

	t.Run("missing name", func(t *testing.T) {
		_, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
			Icon:      "💼",
			MemberIDs: []uuid.UUID{memberA, memberB},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("missing icon", func(t *testing.T) {
		_, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
			Name:      "Test",
			MemberIDs: []uuid.UUID{memberA, memberB},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "icon")
	})

	t.Run("too few members", func(t *testing.T) {
		_, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
			Name:      "Test",
			Icon:      "💼",
			MemberIDs: []uuid.UUID{memberA},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 2")
	})

	t.Run("creator in member list is excluded", func(t *testing.T) {
		_, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
			Name:      "Test",
			Icon:      "💼",
			MemberIDs: []uuid.UUID{creator, memberA},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 2")
	})

	t.Run("non-existent member", func(t *testing.T) {
		_, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
			Name:      "Test",
			Icon:      "💼",
			MemberIDs: []uuid.UUID{memberA, uuid.New()},
		})
		require.Error(t, err)
	})
}

func TestGroupService_UpdateGroup(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub)

	creator := uuid.New()
	memberA := uuid.New()
	memberB := uuid.New()
	userRepo.addUser(&model.User{ID: creator, Phone: "+628111", Name: "Creator"})
	userRepo.addUser(&model.User{ID: memberA, Phone: "+628222", Name: "MemberA"})
	userRepo.addUser(&model.User{ID: memberB, Phone: "+628333", Name: "MemberB"})

	group, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
		Name:      "Original",
		Icon:      "💼",
		MemberIDs: []uuid.UUID{memberA, memberB},
	})
	require.NoError(t, err)

	t.Run("admin can update", func(t *testing.T) {
		newName := "Updated Name"
		updated, err := svc.UpdateGroup(context.Background(), group.ID, creator, UpdateGroupInput{
			Name: &newName,
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
	})

	t.Run("non-admin cannot update", func(t *testing.T) {
		name := "Hacked"
		_, err := svc.UpdateGroup(context.Background(), group.ID, memberA, UpdateGroupInput{
			Name: &name,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "admin")
	})
}

func TestGroupService_AddMember(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub)

	creator := uuid.New()
	memberA := uuid.New()
	memberB := uuid.New()
	memberC := uuid.New()
	userRepo.addUser(&model.User{ID: creator, Phone: "+628111", Name: "Creator"})
	userRepo.addUser(&model.User{ID: memberA, Phone: "+628222", Name: "MemberA"})
	userRepo.addUser(&model.User{ID: memberB, Phone: "+628333", Name: "MemberB"})
	userRepo.addUser(&model.User{ID: memberC, Phone: "+628444", Name: "MemberC"})

	group, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
		Name:      "Test Group",
		Icon:      "💼",
		MemberIDs: []uuid.UUID{memberA, memberB},
	})
	require.NoError(t, err)

	t.Run("admin can add member", func(t *testing.T) {
		err := svc.AddMember(context.Background(), group.ID, memberC, creator)
		require.NoError(t, err)

		members, _ := chatRepo.GetMembers(context.Background(), group.ID)
		assert.Len(t, members, 4)
	})

	t.Run("non-admin cannot add member", func(t *testing.T) {
		newUser := uuid.New()
		userRepo.addUser(&model.User{ID: newUser, Phone: "+628555", Name: "New"})
		err := svc.AddMember(context.Background(), group.ID, newUser, memberA)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "admin")
	})

	t.Run("cannot add existing member", func(t *testing.T) {
		err := svc.AddMember(context.Background(), group.ID, memberA, creator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already a member")
	})
}

func TestGroupService_RemoveMember(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub)

	creator := uuid.New()
	memberA := uuid.New()
	memberB := uuid.New()
	userRepo.addUser(&model.User{ID: creator, Phone: "+628111", Name: "Creator"})
	userRepo.addUser(&model.User{ID: memberA, Phone: "+628222", Name: "MemberA"})
	userRepo.addUser(&model.User{ID: memberB, Phone: "+628333", Name: "MemberB"})

	group, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
		Name:      "Test Group",
		Icon:      "💼",
		MemberIDs: []uuid.UUID{memberA, memberB},
	})
	require.NoError(t, err)

	t.Run("admin can remove member", func(t *testing.T) {
		err := svc.RemoveMember(context.Background(), group.ID, memberB, creator)
		require.NoError(t, err)

		members, _ := chatRepo.GetMembers(context.Background(), group.ID)
		assert.Len(t, members, 2) // creator + memberA
	})

	t.Run("cannot remove creator", func(t *testing.T) {
		err := svc.RemoveMember(context.Background(), group.ID, creator, creator)
		require.Error(t, err)
	})

	t.Run("non-admin cannot remove", func(t *testing.T) {
		// memberA is not admin
		err := svc.RemoveMember(context.Background(), group.ID, creator, memberA)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "admin")
	})
}

func TestGroupService_PromoteToAdmin(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub)

	creator := uuid.New()
	memberA := uuid.New()
	memberB := uuid.New()
	userRepo.addUser(&model.User{ID: creator, Phone: "+628111", Name: "Creator"})
	userRepo.addUser(&model.User{ID: memberA, Phone: "+628222", Name: "MemberA"})
	userRepo.addUser(&model.User{ID: memberB, Phone: "+628333", Name: "MemberB"})

	group, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
		Name:      "Test Group",
		Icon:      "💼",
		MemberIDs: []uuid.UUID{memberA, memberB},
	})
	require.NoError(t, err)

	t.Run("promote member to admin", func(t *testing.T) {
		err := svc.PromoteToAdmin(context.Background(), group.ID, memberA, creator)
		require.NoError(t, err)

		members, _ := chatRepo.GetMembers(context.Background(), group.ID)
		for _, m := range members {
			if m.UserID == memberA {
				assert.Equal(t, model.MemberRoleAdmin, m.Role)
			}
		}
	})

	t.Run("already admin", func(t *testing.T) {
		err := svc.PromoteToAdmin(context.Background(), group.ID, memberA, creator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already an admin")
	})

	t.Run("non-admin cannot promote", func(t *testing.T) {
		err := svc.PromoteToAdmin(context.Background(), group.ID, memberB, memberB)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "admin")
	})
}

func TestGroupService_LeaveGroup(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub)

	creator := uuid.New()
	memberA := uuid.New()
	memberB := uuid.New()
	userRepo.addUser(&model.User{ID: creator, Phone: "+628111", Name: "Creator"})
	userRepo.addUser(&model.User{ID: memberA, Phone: "+628222", Name: "MemberA"})
	userRepo.addUser(&model.User{ID: memberB, Phone: "+628333", Name: "MemberB"})

	group, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
		Name:      "Test Group",
		Icon:      "💼",
		MemberIDs: []uuid.UUID{memberA, memberB},
	})
	require.NoError(t, err)

	t.Run("member can leave", func(t *testing.T) {
		err := svc.LeaveGroup(context.Background(), group.ID, memberA)
		require.NoError(t, err)

		members, _ := chatRepo.GetMembers(context.Background(), group.ID)
		assert.Len(t, members, 2) // creator + memberB
	})

	t.Run("creator cannot leave", func(t *testing.T) {
		err := svc.LeaveGroup(context.Background(), group.ID, creator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "creator")
	})

	t.Run("non-member cannot leave", func(t *testing.T) {
		err := svc.LeaveGroup(context.Background(), group.ID, uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a member")
	})
}

func TestGroupService_DeleteGroup(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub)

	creator := uuid.New()
	memberA := uuid.New()
	memberB := uuid.New()
	userRepo.addUser(&model.User{ID: creator, Phone: "+628111", Name: "Creator"})
	userRepo.addUser(&model.User{ID: memberA, Phone: "+628222", Name: "MemberA"})
	userRepo.addUser(&model.User{ID: memberB, Phone: "+628333", Name: "MemberB"})

	group, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
		Name:      "Test Group",
		Icon:      "💼",
		MemberIDs: []uuid.UUID{memberA, memberB},
	})
	require.NoError(t, err)

	t.Run("non-creator cannot delete", func(t *testing.T) {
		err := svc.DeleteGroup(context.Background(), group.ID, memberA)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "creator")
	})

	t.Run("creator can delete", func(t *testing.T) {
		err := svc.DeleteGroup(context.Background(), group.ID, creator)
		require.NoError(t, err)

		// Chat should be removed
		_, ok := chatRepo.chats[group.ID]
		assert.False(t, ok)
	})
}

func TestGroupService_GetGroupInfo(t *testing.T) {
	chatRepo := newMockChatRepo()
	msgRepo := newMockMessageRepo()
	msgStatRepo := newMockMessageStatRepo()
	userRepo := newMockUserRepo()
	hub := newTestHub()
	defer hub.Shutdown()

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub)

	creator := uuid.New()
	memberA := uuid.New()
	memberB := uuid.New()
	userRepo.addUser(&model.User{ID: creator, Phone: "+628111", Name: "Creator"})
	userRepo.addUser(&model.User{ID: memberA, Phone: "+628222", Name: "MemberA"})
	userRepo.addUser(&model.User{ID: memberB, Phone: "+628333", Name: "MemberB"})

	group, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
		Name:      "Info Group",
		Icon:      "🎯",
		MemberIDs: []uuid.UUID{memberA, memberB},
	})
	require.NoError(t, err)

	t.Run("member can get info", func(t *testing.T) {
		info, err := svc.GetGroupInfo(context.Background(), group.ID, memberA)
		require.NoError(t, err)
		assert.Equal(t, "Info Group", info.Chat.Name)
		assert.Len(t, info.Members, 3)
	})

	t.Run("non-member cannot get info", func(t *testing.T) {
		_, err := svc.GetGroupInfo(context.Background(), group.ID, uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a member")
	})
}
