package handler_test

import (
	"context"

	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/service"
)

// --- Mock ChatService ---

type mockChatService struct {
	chatList   []*service.ChatListItem
	chatDetail *service.ChatDetail
	chat       *model.Chat
	isMember   bool
	err        error
}

func (m *mockChatService) CreatePersonalChat(_ context.Context, _, _ uuid.UUID) (*model.Chat, error) {
	return m.chat, m.err
}

func (m *mockChatService) GetOrCreatePersonalChat(_ context.Context, _, _ uuid.UUID) (*model.Chat, error) {
	return m.chat, m.err
}

func (m *mockChatService) ListChats(_ context.Context, _ uuid.UUID) ([]*service.ChatListItem, error) {
	return m.chatList, m.err
}

func (m *mockChatService) GetChat(_ context.Context, _, _ uuid.UUID) (*service.ChatDetail, error) {
	return m.chatDetail, m.err
}

func (m *mockChatService) PinChat(_ context.Context, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockChatService) UnpinChat(_ context.Context, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockChatService) IsMember(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return m.isMember, m.err
}

// --- Mock MessageService ---

type mockMessageService struct {
	message     *model.Message
	messagePage *service.MessagePage
	messages    []*model.Message
	err         error
}

func (m *mockMessageService) SendMessage(_ context.Context, _ service.SendMessageInput) (*model.Message, error) {
	return m.message, m.err
}

func (m *mockMessageService) GetMessages(_ context.Context, _ uuid.UUID, _ string, _ int) (*service.MessagePage, error) {
	return m.messagePage, m.err
}

func (m *mockMessageService) ForwardMessage(_ context.Context, _, _, _ uuid.UUID) (*model.Message, error) {
	return m.message, m.err
}

func (m *mockMessageService) DeleteMessage(_ context.Context, _, _ uuid.UUID, _ bool) error {
	return m.err
}

func (m *mockMessageService) SearchMessages(_ context.Context, _ uuid.UUID, _ string) ([]*model.Message, error) {
	return m.messages, m.err
}

func (m *mockMessageService) MarkChatAsRead(_ context.Context, _, _ uuid.UUID) error {
	return m.err
}

// --- Mock GroupService ---

type mockGroupService struct {
	chat      *model.Chat
	groupInfo *service.GroupInfo
	err       error
}

func (m *mockGroupService) CreateGroup(_ context.Context, _ uuid.UUID, _ service.CreateGroupInput) (*model.Chat, error) {
	return m.chat, m.err
}

func (m *mockGroupService) UpdateGroup(_ context.Context, _, _ uuid.UUID, _ service.UpdateGroupInput) (*model.Chat, error) {
	return m.chat, m.err
}

func (m *mockGroupService) AddMember(_ context.Context, _, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockGroupService) RemoveMember(_ context.Context, _, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockGroupService) PromoteToAdmin(_ context.Context, _, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockGroupService) LeaveGroup(_ context.Context, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockGroupService) DeleteGroup(_ context.Context, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockGroupService) GetGroupInfo(_ context.Context, _, _ uuid.UUID) (*service.GroupInfo, error) {
	return m.groupInfo, m.err
}

// --- Mock SearchService ---

type mockSearchService struct {
	searchResults *service.SearchResults
	msgRows       []*model.MessageSearchRow
	docRows       []*model.DocumentSearchRow
	contacts      []*model.User
	entities      []*model.Entity
	err           error
}

func (m *mockSearchService) SearchAll(_ context.Context, _ uuid.UUID, _ string, _ int) (*service.SearchResults, error) {
	return m.searchResults, m.err
}

func (m *mockSearchService) SearchMessages(_ context.Context, _ uuid.UUID, _ string, _ service.SearchOpts) ([]*model.MessageSearchRow, error) {
	return m.msgRows, m.err
}

func (m *mockSearchService) SearchDocuments(_ context.Context, _ uuid.UUID, _ string, _ service.SearchOpts) ([]*model.DocumentSearchRow, error) {
	return m.docRows, m.err
}

func (m *mockSearchService) SearchContacts(_ context.Context, _ uuid.UUID, _ string) ([]*model.User, error) {
	return m.contacts, m.err
}

func (m *mockSearchService) SearchEntities(_ context.Context, _ uuid.UUID, _ string) ([]*model.Entity, error) {
	return m.entities, m.err
}

func (m *mockSearchService) SearchInChat(_ context.Context, _, _ uuid.UUID, _ string, _ service.SearchOpts) ([]*model.MessageSearchRow, error) {
	return m.msgRows, m.err
}

// --- Mock BackupService ---

type mockBackupService struct {
	bundle  *model.BackupBundle
	record  *model.BackupRecord
	records []model.BackupRecord
	err     error
}

func (m *mockBackupService) ExportUserData(_ context.Context, _ uuid.UUID) (*model.BackupBundle, error) {
	return m.bundle, m.err
}

func (m *mockBackupService) ImportUserData(_ context.Context, _ uuid.UUID, _ *model.BackupBundle) error {
	return m.err
}

func (m *mockBackupService) LogBackup(_ context.Context, _ uuid.UUID, _ model.LogBackupInput) (*model.BackupRecord, error) {
	return m.record, m.err
}

func (m *mockBackupService) GetBackupHistory(_ context.Context, _ uuid.UUID) ([]model.BackupRecord, error) {
	return m.records, m.err
}

func (m *mockBackupService) GetLatestBackup(_ context.Context, _ uuid.UUID) (*model.BackupRecord, error) {
	return m.record, m.err
}

// --- Mock NotificationService ---

type mockNotificationService struct {
	err error
}

func (m *mockNotificationService) RegisterDevice(_ context.Context, _ uuid.UUID, _, _ string) error {
	return m.err
}

func (m *mockNotificationService) UnregisterDevice(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockNotificationService) SendToUser(_ context.Context, _ uuid.UUID, _ model.Notification) error {
	return m.err
}

func (m *mockNotificationService) SendToUsers(_ context.Context, _ []uuid.UUID, _ model.Notification) error {
	return m.err
}

func (m *mockNotificationService) SendToChat(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ model.Notification) error {
	return m.err
}

// --- Mock EntityService ---

type mockEntityService struct {
	entity    *model.Entity
	entities  []*model.Entity
	listItems []*model.EntityListItem
	total     int
	types     []string
	docs      []*model.Document
	err       error
}

func (m *mockEntityService) Create(_ context.Context, _ uuid.UUID, _ service.CreateEntityInput) (*model.Entity, error) {
	return m.entity, m.err
}

func (m *mockEntityService) GetByID(_ context.Context, _, _ uuid.UUID) (*model.Entity, error) {
	return m.entity, m.err
}

func (m *mockEntityService) List(_ context.Context, _ uuid.UUID, _ string, _, _ int) ([]*model.EntityListItem, int, error) {
	return m.listItems, m.total, m.err
}

func (m *mockEntityService) Update(_ context.Context, _, _ uuid.UUID, _ service.UpdateEntityInput) (*model.Entity, error) {
	return m.entity, m.err
}

func (m *mockEntityService) Delete(_ context.Context, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockEntityService) Search(_ context.Context, _ uuid.UUID, _ string) ([]*model.Entity, error) {
	return m.entities, m.err
}

func (m *mockEntityService) ListTypes(_ context.Context, _ uuid.UUID) ([]string, error) {
	return m.types, m.err
}

func (m *mockEntityService) LinkToDocument(_ context.Context, _, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockEntityService) UnlinkFromDocument(_ context.Context, _, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockEntityService) GetDocumentEntities(_ context.Context, _ uuid.UUID) ([]*model.Entity, error) {
	return m.entities, m.err
}

func (m *mockEntityService) GetEntityDocuments(_ context.Context, _ uuid.UUID) ([]*model.Document, error) {
	return m.docs, m.err
}

func (m *mockEntityService) CreateFromContact(_ context.Context, _, _ uuid.UUID) (*model.Entity, error) {
	return m.entity, m.err
}

// --- Mock MediaService ---

type mockMediaService struct {
	mediaResp   *model.MediaResponse
	downloadURL string
	err         error
}

func (m *mockMediaService) Upload(_ context.Context, _ service.MediaUploadInput) (*model.MediaResponse, error) {
	return m.mediaResp, m.err
}

func (m *mockMediaService) GetByID(_ context.Context, _ uuid.UUID) (*model.MediaResponse, error) {
	return m.mediaResp, m.err
}

func (m *mockMediaService) GetDownloadURL(_ context.Context, _ uuid.UUID) (string, error) {
	return m.downloadURL, m.err
}

func (m *mockMediaService) Delete(_ context.Context, _, _ uuid.UUID) error {
	return m.err
}

// --- Mock DocumentService ---

type mockDocumentService struct {
	docFull *service.DocumentFull
	doc     *model.Document
	docList []*service.DocumentListItem
	signers []*model.DocumentSigner
	history []*model.DocumentHistory
	err     error
}

func (m *mockDocumentService) Create(_ context.Context, _ service.CreateDocumentInput) (*service.DocumentFull, error) {
	return m.docFull, m.err
}

func (m *mockDocumentService) GetByID(_ context.Context, _, _ uuid.UUID) (*service.DocumentFull, error) {
	return m.docFull, m.err
}

func (m *mockDocumentService) ListByContext(_ context.Context, _ string, _ uuid.UUID) ([]*service.DocumentListItem, error) {
	return m.docList, m.err
}

func (m *mockDocumentService) ListAll(_ context.Context, _ uuid.UUID) ([]*service.DocumentListItem, error) {
	return m.docList, m.err
}

func (m *mockDocumentService) Update(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ model.UpdateDocumentInput) (*model.Document, error) {
	return m.doc, m.err
}

func (m *mockDocumentService) Delete(_ context.Context, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockDocumentService) Duplicate(_ context.Context, _, _ uuid.UUID) (*service.DocumentFull, error) {
	return m.docFull, m.err
}

func (m *mockDocumentService) AddCollaborator(_ context.Context, _, _, _ uuid.UUID, _ model.CollaboratorRole) error {
	return m.err
}

func (m *mockDocumentService) RemoveCollaborator(_ context.Context, _, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockDocumentService) UpdateCollaboratorRole(_ context.Context, _, _, _ uuid.UUID, _ model.CollaboratorRole) error {
	return m.err
}

func (m *mockDocumentService) AddTag(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockDocumentService) RemoveTag(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockDocumentService) GetHistory(_ context.Context, _ uuid.UUID) ([]*model.DocumentHistory, error) {
	return m.history, m.err
}

func (m *mockDocumentService) LockDocument(_ context.Context, _, _ uuid.UUID, _ model.LockedByType) error {
	return m.err
}

func (m *mockDocumentService) UnlockDocument(_ context.Context, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockDocumentService) AddSigner(_ context.Context, _, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockDocumentService) RemoveSigner(_ context.Context, _, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockDocumentService) SignDocument(_ context.Context, _, _ uuid.UUID, _ string) (*model.Document, error) {
	return m.doc, m.err
}

func (m *mockDocumentService) ListSigners(_ context.Context, _ uuid.UUID) ([]*model.DocumentSigner, error) {
	return m.signers, m.err
}

// --- Mock BlockService ---

type mockBlockService struct {
	block  *model.Block
	blocks []*model.Block
	err    error
}

func (m *mockBlockService) AddBlock(_ context.Context, _, _ uuid.UUID, _ service.AddBlockInput) (*model.Block, error) {
	return m.block, m.err
}

func (m *mockBlockService) UpdateBlock(_ context.Context, _, _ uuid.UUID, _ model.UpdateBlockInput) (*model.Block, error) {
	return m.block, m.err
}

func (m *mockBlockService) DeleteBlock(_ context.Context, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockBlockService) MoveBlock(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int) error {
	return m.err
}

func (m *mockBlockService) GetBlocks(_ context.Context, _ uuid.UUID) ([]*model.Block, error) {
	return m.blocks, m.err
}

func (m *mockBlockService) ReorderBlocks(_ context.Context, _, _ uuid.UUID, _ []uuid.UUID) error {
	return m.err
}

func (m *mockBlockService) BatchUpdate(_ context.Context, _, _ uuid.UUID, _ []service.BlockOperation) error {
	return m.err
}

// --- Mock TemplateService ---

type mockTemplateService struct {
	templates []*service.DocumentTemplate
	template  *service.DocumentTemplate
	blocks    []service.TemplateBlock
}

func (m *mockTemplateService) GetTemplates() []*service.DocumentTemplate {
	return m.templates
}

func (m *mockTemplateService) GetTemplate(id string) *service.DocumentTemplate {
	return m.template
}

func (m *mockTemplateService) GetTemplateBlocks(id string) []service.TemplateBlock {
	return m.blocks
}

// --- Mock TopicService ---

type mockTopicService struct {
	topic     *model.Topic
	topicDtl  *service.TopicDetail
	topicList []*service.TopicListItem
	err       error
}

func (m *mockTopicService) CreateTopic(_ context.Context, _ uuid.UUID, _ service.CreateTopicInput) (*model.Topic, error) {
	return m.topic, m.err
}

func (m *mockTopicService) GetTopic(_ context.Context, _, _ uuid.UUID) (*service.TopicDetail, error) {
	return m.topicDtl, m.err
}

func (m *mockTopicService) ListByChat(_ context.Context, _, _ uuid.UUID) ([]*service.TopicListItem, error) {
	return m.topicList, m.err
}

func (m *mockTopicService) ListByUser(_ context.Context, _ uuid.UUID) ([]*service.TopicListItem, error) {
	return m.topicList, m.err
}

func (m *mockTopicService) UpdateTopic(_ context.Context, _, _ uuid.UUID, _ service.UpdateTopicInput) (*model.Topic, error) {
	return m.topic, m.err
}

func (m *mockTopicService) AddMember(_ context.Context, _, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockTopicService) RemoveMember(_ context.Context, _, _, _ uuid.UUID) error {
	return m.err
}

func (m *mockTopicService) DeleteTopic(_ context.Context, _, _ uuid.UUID) error {
	return m.err
}

// --- Mock TopicMessageService ---

type mockTopicMessageService struct {
	message     *model.TopicMessage
	messagePage *service.TopicMessagePage
	err         error
}

func (m *mockTopicMessageService) SendMessage(_ context.Context, _ service.SendTopicMessageInput) (*model.TopicMessage, error) {
	return m.message, m.err
}

func (m *mockTopicMessageService) GetMessages(_ context.Context, _ uuid.UUID, _ string, _ int) (*service.TopicMessagePage, error) {
	return m.messagePage, m.err
}

func (m *mockTopicMessageService) DeleteMessage(_ context.Context, _, _ uuid.UUID, _ bool) error {
	return m.err
}

// --- Mock OTPService ---

type mockOTPService struct {
	code string
	err  error
}

func (m *mockOTPService) Generate(_ context.Context, _ string) (string, error) {
	return m.code, m.err
}

func (m *mockOTPService) Verify(_ context.Context, _, _ string) error {
	return m.err
}

// --- Mock ReverseOTPService ---

type mockReverseOTPService struct {
	session *service.ReverseOTPSession
	result  *service.VerificationResult
	err     error
}

func (m *mockReverseOTPService) InitSession(_ context.Context, _ string) (*service.ReverseOTPSession, error) {
	return m.session, m.err
}

func (m *mockReverseOTPService) CheckVerification(_ context.Context, _ string) (*service.VerificationResult, error) {
	return m.result, m.err
}

func (m *mockReverseOTPService) HandleIncomingMessage(_ context.Context, _, _ string) error {
	return m.err
}

// --- Mock TokenService ---

type mockTokenService struct {
	tokenPair *service.TokenPair
	claims    *service.Claims
	err       error
}

func (m *mockTokenService) Generate(_ context.Context, _ uuid.UUID) (*service.TokenPair, error) {
	return m.tokenPair, m.err
}

func (m *mockTokenService) Validate(_ string) (*service.Claims, error) {
	return m.claims, m.err
}

func (m *mockTokenService) Refresh(_ context.Context, _ string) (*service.TokenPair, error) {
	return m.tokenPair, m.err
}

func (m *mockTokenService) Revoke(_ context.Context, _, _ string) error {
	return m.err
}

// --- Mock SessionService ---

type mockSessionService struct {
	err error
}

func (m *mockSessionService) Register(_ context.Context, _ uuid.UUID, _, _ string) error {
	return m.err
}

func (m *mockSessionService) Validate(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockSessionService) Invalidate(_ context.Context, _ uuid.UUID) error {
	return m.err
}

// --- Mock UserRepository ---

type mockUserRepo struct {
	user         *model.User
	users        []*model.User
	err          error
	findPhoneErr *error // optional: override err for FindByPhone
}

func (m *mockUserRepo) FindByPhone(_ context.Context, _ string) (*model.User, error) {
	if m.findPhoneErr != nil {
		return nil, *m.findPhoneErr
	}
	return m.user, m.err
}

func (m *mockUserRepo) FindByID(_ context.Context, _ uuid.UUID) (*model.User, error) {
	return m.user, m.err
}

func (m *mockUserRepo) Create(_ context.Context, _ model.CreateUserInput) (*model.User, error) {
	return m.user, m.err
}

func (m *mockUserRepo) FindByPhones(_ context.Context, _ []string) ([]*model.User, error) {
	return m.users, m.err
}

func (m *mockUserRepo) FindByPhoneHashes(_ context.Context, _ []string) ([]*model.User, error) {
	return m.users, m.err
}

func (m *mockUserRepo) Update(_ context.Context, _ uuid.UUID, _ model.UpdateUserInput) (*model.User, error) {
	return m.user, m.err
}

func (m *mockUserRepo) UpdatePhoneHash(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockUserRepo) UpdateLastSeen(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func (m *mockUserRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return m.err
}

// ensure mockUserRepo implements repository.UserRepository
var _ repository.UserRepository = (*mockUserRepo)(nil)
