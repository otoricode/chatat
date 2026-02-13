package service

import (
	"context"
	"fmt"
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

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub, nil)

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

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub, nil)

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

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub, nil)

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

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub, nil)

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

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub, nil)

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

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub, nil)

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

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub, nil)

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

	svc := NewGroupService(chatRepo, msgRepo, msgStatRepo, userRepo, hub, nil)

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

func TestGroupService_CreateGroup_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("name too long", func(t *testing.T) {
		svc := NewGroupService(newMockChatRepo(), newMockMessageRepo(), newMockMessageStatRepo(), newMockUserRepo(), hub, nil)
		longName := ""
		for i := 0; i < 101; i++ {
			longName += "a"
		}
		_, err := svc.CreateGroup(context.Background(), uuid.New(), CreateGroupInput{
			Name: longName, Icon: "x", MemberIDs: []uuid.UUID{uuid.New(), uuid.New()},
		})
		require.Error(t, err)
	})

	t.Run("verify member generic error", func(t *testing.T) {
		userRepo := newMockUserRepo()
		userRepo.createErr = fmt.Errorf("db") // won't help - FindByID uses map
		svc := NewGroupService(newMockChatRepo(), newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub, nil)
		// memberIDs are unknown UUIDs → FindByID returns NotFound
		_, err := svc.CreateGroup(context.Background(), uuid.New(), CreateGroupInput{
			Name: "G", Icon: "x", MemberIDs: []uuid.UUID{uuid.New(), uuid.New()},
		})
		require.Error(t, err)
	})

	t.Run("create chat error", func(t *testing.T) {
		userRepo := newMockUserRepo()
		chatRepo := newMockChatRepo()
		a, b := uuid.New(), uuid.New()
		userRepo.addUser(&model.User{ID: a, Phone: "+1", Name: "A"})
		userRepo.addUser(&model.User{ID: b, Phone: "+2", Name: "B"})
		chatRepo.createErr = fmt.Errorf("db error")
		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub, nil)
		_, err := svc.CreateGroup(context.Background(), uuid.New(), CreateGroupInput{
			Name: "G", Icon: "x", MemberIDs: []uuid.UUID{a, b},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create group")
	})

	t.Run("add creator error", func(t *testing.T) {
		userRepo := newMockUserRepo()
		chatRepo := newMockChatRepo()
		a, b := uuid.New(), uuid.New()
		userRepo.addUser(&model.User{ID: a, Phone: "+1", Name: "A"})
		userRepo.addUser(&model.User{ID: b, Phone: "+2", Name: "B"})
		chatRepo.addMemberErr = fmt.Errorf("db error")
		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub, nil)
		_, err := svc.CreateGroup(context.Background(), uuid.New(), CreateGroupInput{
			Name: "G", Icon: "x", MemberIDs: []uuid.UUID{a, b},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "add creator")
	})
}

func TestGroupService_UpdateGroup_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("find chat error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		creator := uuid.New()
		a, b := uuid.New(), uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})
		userRepo.addUser(&model.User{ID: a, Phone: "+2", Name: "A"})
		userRepo.addUser(&model.User{ID: b, Phone: "+3", Name: "B"})

		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub, nil)
		group, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
			Name: "G", Icon: "x", MemberIDs: []uuid.UUID{a, b},
		})
		require.NoError(t, err)

		chatRepo.findErr = fmt.Errorf("db error")
		name := "New"
		_, err = svc.UpdateGroup(context.Background(), group.ID, creator, UpdateGroupInput{Name: &name})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find chat")
	})

	t.Run("not a group", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		creator := uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})

		// Create a personal chat
		personalChat := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: creator}
		chatRepo.chats[personalChat.ID] = personalChat
		_ = chatRepo.AddMember(context.Background(), personalChat.ID, creator, model.MemberRoleAdmin)

		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub, nil)
		name := "New"
		_, err := svc.UpdateGroup(context.Background(), personalChat.ID, creator, UpdateGroupInput{Name: &name})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "group chats")
	})
}

func TestGroupService_AddMember_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("find chat error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		creator := uuid.New()
		a, b := uuid.New(), uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})
		userRepo.addUser(&model.User{ID: a, Phone: "+2", Name: "A"})
		userRepo.addUser(&model.User{ID: b, Phone: "+3", Name: "B"})

		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub, nil)
		group, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
			Name: "G", Icon: "x", MemberIDs: []uuid.UUID{a, b},
		})
		require.NoError(t, err)

		chatRepo.findErr = fmt.Errorf("db error")
		newMember := uuid.New()
		userRepo.addUser(&model.User{ID: newMember, Phone: "+4", Name: "New"})
		err = svc.AddMember(context.Background(), group.ID, newMember, creator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find chat")
	})

	t.Run("add to personal chat", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		creator := uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})

		personalChat := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: creator}
		chatRepo.chats[personalChat.ID] = personalChat
		_ = chatRepo.AddMember(context.Background(), personalChat.ID, creator, model.MemberRoleAdmin)

		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub, nil)
		err := svc.AddMember(context.Background(), personalChat.ID, uuid.New(), creator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "group chats")
	})

	t.Run("user not found", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		creator := uuid.New()
		a, b := uuid.New(), uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})
		userRepo.addUser(&model.User{ID: a, Phone: "+2", Name: "A"})
		userRepo.addUser(&model.User{ID: b, Phone: "+3", Name: "B"})

		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub, nil)
		group, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
			Name: "G", Icon: "x", MemberIDs: []uuid.UUID{a, b},
		})
		require.NoError(t, err)

		err = svc.AddMember(context.Background(), group.ID, uuid.New(), creator) // unknown user
		require.Error(t, err)
	})
}

func TestGroupService_RemoveMember_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("remove self", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		creator := uuid.New()
		a, b := uuid.New(), uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})
		userRepo.addUser(&model.User{ID: a, Phone: "+2", Name: "A"})
		userRepo.addUser(&model.User{ID: b, Phone: "+3", Name: "B"})

		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub, nil)
		group, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
			Name: "G", Icon: "x", MemberIDs: []uuid.UUID{a, b},
		})
		require.NoError(t, err)

		err = svc.RemoveMember(context.Background(), group.ID, creator, creator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "leave")
	})

	t.Run("find chat error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		creator := uuid.New()
		a, b := uuid.New(), uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})
		userRepo.addUser(&model.User{ID: a, Phone: "+2", Name: "A"})
		userRepo.addUser(&model.User{ID: b, Phone: "+3", Name: "B"})

		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub, nil)
		group, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
			Name: "G", Icon: "x", MemberIDs: []uuid.UUID{a, b},
		})
		require.NoError(t, err)

		chatRepo.findErr = fmt.Errorf("db error")
		err = svc.RemoveMember(context.Background(), group.ID, a, creator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find chat")
	})
}

func TestGroupService_LeaveGroup_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("get members error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		chatRepo.getMembersErr = fmt.Errorf("db error")
		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), newMockUserRepo(), hub, nil)
		err := svc.LeaveGroup(context.Background(), uuid.New(), uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get members")
	})
}

func TestGroupService_DeleteGroup_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("find chat error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		chatRepo.findErr = fmt.Errorf("db error")
		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), newMockUserRepo(), hub, nil)
		err := svc.DeleteGroup(context.Background(), uuid.New(), uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find chat")
	})

	t.Run("delete personal chat", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		creator := uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})

		personalChat := &model.Chat{ID: uuid.New(), Type: model.ChatTypePersonal, CreatedBy: creator}
		chatRepo.chats[personalChat.ID] = personalChat

		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub, nil)
		err := svc.DeleteGroup(context.Background(), personalChat.ID, creator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "group chats")
	})
}

func TestGroupService_GetGroupInfo_Errors(t *testing.T) {
	hub := newTestHub()
	defer hub.Shutdown()

	t.Run("get members error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		chatRepo.getMembersErr = fmt.Errorf("db error")
		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), newMockUserRepo(), hub, nil)
		_, err := svc.GetGroupInfo(context.Background(), uuid.New(), uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get members")
	})

	t.Run("find chat error", func(t *testing.T) {
		chatRepo := newMockChatRepo()
		userRepo := newMockUserRepo()
		creator := uuid.New()
		a, b := uuid.New(), uuid.New()
		userRepo.addUser(&model.User{ID: creator, Phone: "+1", Name: "C"})
		userRepo.addUser(&model.User{ID: a, Phone: "+2", Name: "A"})
		userRepo.addUser(&model.User{ID: b, Phone: "+3", Name: "B"})

		svc := NewGroupService(chatRepo, newMockMessageRepo(), newMockMessageStatRepo(), userRepo, hub, nil)
		group, err := svc.CreateGroup(context.Background(), creator, CreateGroupInput{
			Name: "G", Icon: "x", MemberIDs: []uuid.UUID{a, b},
		})
		require.NoError(t, err)

		chatRepo.findErr = fmt.Errorf("db error")
		_, err = svc.GetGroupInfo(context.Background(), group.ID, creator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find chat")
	})
}
