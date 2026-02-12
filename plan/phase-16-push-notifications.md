# Phase 16: Push Notifications

> Implementasi push notification untuk pesan baru, permintaan tanda tangan,
> dan event penting lainnya. Support FCM (Android) dan APNs (iOS).

**Estimasi:** 3 hari
**Dependency:** Phase 04 (Auth), Phase 07 (Chat), Phase 14 (Doc Collab)
**Output:** Push notification system yang reliable dengan deep linking.

---

## Task 16.1: Notification Backend Service

**Input:** FCM/APNs credentials, user device tokens
**Output:** Notification dispatch service

### Steps:
1. Buat `internal/service/notification_service.go`:
   ```go
   type NotificationService interface {
       RegisterDevice(ctx context.Context, userID uuid.UUID, input RegisterDeviceInput) error
       UnregisterDevice(ctx context.Context, userID uuid.UUID, deviceToken string) error
       SendToUser(ctx context.Context, userID uuid.UUID, notif Notification) error
       SendToUsers(ctx context.Context, userIDs []uuid.UUID, notif Notification) error
       SendToChat(ctx context.Context, chatID uuid.UUID, excludeUserID uuid.UUID, notif Notification) error
   }

   type RegisterDeviceInput struct {
       Token    string `json:"token" validate:"required"`
       Platform string `json:"platform" validate:"required,oneof=ios android"`
   }

   type Notification struct {
       Type     string            `json:"type"` // message, signature_request, group_invite, document_locked
       Title    string            `json:"title"`
       Body     string            `json:"body"`
       Data     map[string]string `json:"data"` // deep link data
       Badge    int               `json:"badge"`
       Sound    string            `json:"sound"`
       Priority string            `json:"priority"` // high, normal
   }
   ```
2. Device token storage:
   ```go
   // Migration: device_tokens table
   CREATE TABLE device_tokens (
       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
       user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
       token TEXT NOT NULL,
       platform TEXT NOT NULL, -- ios, android
       created_at TIMESTAMPTZ DEFAULT NOW(),
       updated_at TIMESTAMPTZ DEFAULT NOW(),
       UNIQUE(user_id, token)
   );
   ```
3. FCM integration (via `firebase.google.com/go/v4/messaging`):
   ```go
   type FCMSender struct {
       client *messaging.Client
   }

   func (s *FCMSender) Send(ctx context.Context, token string, notif Notification) error {
       msg := &messaging.Message{
           Token: token,
           Notification: &messaging.Notification{
               Title: notif.Title,
               Body:  notif.Body,
           },
           Data: notif.Data,
           Android: &messaging.AndroidConfig{
               Priority: notif.Priority,
           },
           APNS: &messaging.APNSConfig{
               Payload: &messaging.APNSPayload{
                   Aps: &messaging.Aps{
                       Badge: &notif.Badge,
                       Sound: notif.Sound,
                   },
               },
           },
       }
       _, err := s.client.Send(ctx, msg)
       return err
   }
   ```
4. Notification types:
   - `message`: "Ahmad: Halo, apa kabar?" → deep link to chat
   - `group_message`: "[Keluarga] Ahmad: Besok kumpul" → deep link to group
   - `topic_message`: "[Keuangan] Ahmad: Sudah bayar" → deep link to topic
   - `signature_request`: "Ahmad meminta tanda tangan Anda" → deep link to document
   - `document_locked`: "Dokumen 'Notulen' telah dikunci" → deep link to document
   - `group_invite`: "Ahmad mengundang Anda ke grup" → deep link to group
5. Handle token expiry:
   - If FCM returns invalid token → remove from DB
   - Periodic cleanup of stale tokens (30 days unused)

### Acceptance Criteria:
- [ ] Device token registration (iOS + Android)
- [ ] FCM message dispatch (both platforms)
- [ ] All notification types defined
- [ ] Deep link data included
- [ ] Invalid token cleanup
- [ ] Send to single user / multiple users / chat members
- [ ] Exclude sender from receiving their own notification

### Testing:
- [ ] Unit test: register/unregister device
- [ ] Unit test: build FCM message per notification type
- [ ] Unit test: invalid token handling
- [ ] Unit test: exclude sender
- [ ] Integration test: send notification (mock FCM)

---

## Task 16.2: Notification Triggers

**Input:** Task 16.1, existing message/document services
**Output:** Auto-send notifications on events

### Steps:
1. Hook notifications into existing services:
   ```go
   // In MessageService.Send():
   // After message saved → SendToChat(chatID, excludeUserID=senderID, notif)

   // In DocumentService.RequestSignatures():
   // For each signer → SendToUser(signerID, signatureRequestNotif)

   // In DocumentService.Lock():
   // Notify all collaborators → SendToUsers(collaboratorIDs, docLockedNotif)

   // In GroupService.AddMember():
   // Notify new member → SendToUser(newMemberID, groupInviteNotif)
   ```
2. Notification formatting (Indonesian):
   - Message: `"[SenderName]: [preview first 50 chars]"`
   - Group message: `"[GroupName] [SenderName]: [preview]"`
   - Topic: `"[TopicName] [SenderName]: [preview]"`
   - Signature: `"[SenderName] meminta tanda tangan Anda untuk '[DocTitle]'"`
   - Locked: `"Dokumen '[DocTitle]' telah dikunci oleh [OwnerName]"`
   - Group invite: `"[InviterName] mengundang Anda ke grup '[GroupName]'"`
3. Mute handling:
   - Check user's mute settings before sending
   - Muted chat → skip notification
   - Consider DND hours (future feature)
4. Badge count:
   - Track unread count per user
   - Update badge on each notification
   - Reset badge on app open

### Acceptance Criteria:
- [ ] New message → notification sent to all chat members
- [ ] Signature request → notification to signer
- [ ] Document locked → notification to collaborators
- [ ] Group invite → notification to invitee
- [ ] Muted chats skip notifications
- [ ] Badge count updated
- [ ] Message preview truncated at 50 chars

### Testing:
- [ ] Unit test: message notification formatting
- [ ] Unit test: muted chat skipped
- [ ] Unit test: badge count increment
- [ ] Integration test: send message → notification triggered

---

## Task 16.3: Mobile Notification Handling

**Input:** Task 16.1, Phase 06 (Mobile Shell)
**Output:** Push notification handling di React Native

### Steps:
1. Setup `@react-native-firebase/messaging`:
   ```typescript
   // src/services/NotificationService.ts
   import messaging from '@react-native-firebase/messaging';

   class NotificationService {
     async initialize(): Promise<void> {
       const authStatus = await messaging().requestPermission();
       const enabled =
         authStatus === messaging.AuthorizationStatus.AUTHORIZED ||
         authStatus === messaging.AuthorizationStatus.PROVISIONAL;

       if (enabled) {
         const token = await messaging().getToken();
         await api.registerDevice({ token, platform: Platform.OS });
       }

       // Listen for token refresh
       messaging().onTokenRefresh(async (newToken) => {
         await api.registerDevice({ token: newToken, platform: Platform.OS });
       });
     }

     setupHandlers(): void {
       // Foreground: show in-app notification
       messaging().onMessage(async (remoteMessage) => {
         this.showInAppNotification(remoteMessage);
       });

       // Background/Killed: handle deep link
       messaging().setBackgroundMessageHandler(async (remoteMessage) => {
         // Update badge, cache message
       });

       // App opened from notification
       messaging().onNotificationOpenedApp((remoteMessage) => {
         this.handleDeepLink(remoteMessage.data);
       });

       // App opened from killed state via notification
       messaging().getInitialNotification().then((remoteMessage) => {
         if (remoteMessage) {
           this.handleDeepLink(remoteMessage.data);
         }
       });
     }
   }
   ```
2. In-app notification (foreground):
   ```typescript
   // Show toast-style notification at top of screen
   const InAppNotification: React.FC<Props> = ({ notification, onPress, onDismiss }) => {
     return (
       <Animated.View style={styles.toast}>
         <TouchableOpacity onPress={onPress} style={styles.toastContent}>
           <Avatar name={notification.senderName} size={36} />
           <View style={styles.toastText}>
             <Text style={styles.toastTitle}>{notification.title}</Text>
             <Text style={styles.toastBody} numberOfLines={1}>{notification.body}</Text>
           </View>
         </TouchableOpacity>
         <TouchableOpacity onPress={onDismiss}>
           <XIcon size={16} color="#9CA3AF" />
         </TouchableOpacity>
       </Animated.View>
     );
   };
   ```
3. Deep linking routes:
   ```typescript
   function handleDeepLink(data: Record<string, string>): void {
     switch (data.type) {
       case 'message':
         navigation.navigate('Chat', { chatId: data.chatId });
         break;
       case 'group_message':
         navigation.navigate('GroupChat', { chatId: data.chatId });
         break;
       case 'topic_message':
         navigation.navigate('Topic', { topicId: data.topicId });
         break;
       case 'signature_request':
         navigation.navigate('DocumentEditor', { docId: data.documentId });
         break;
       case 'document_locked':
         navigation.navigate('DocumentEditor', { docId: data.documentId });
         break;
       case 'group_invite':
         navigation.navigate('GroupChat', { chatId: data.chatId });
         break;
     }
   }
   ```
4. Permission handling:
   - Request on first launch (after auth)
   - Handle denied → show settings prompt
   - iOS provisional notifications support

### Acceptance Criteria:
- [ ] FCM token registered on login
- [ ] Token refreshed on rotate
- [ ] Foreground: in-app toast notification
- [ ] Background: system notification
- [ ] Tap notification → deep link to correct screen
- [ ] Permission request handled gracefully
- [ ] Token unregistered on logout

### Testing:
- [ ] Unit test: deep link routing
- [ ] Component test: InAppNotification renders
- [ ] Component test: in-app notification tap
- [ ] Integration test: permission flow

---

## Phase 16 Review

### Testing Checklist:
- [ ] Device token: register, unregister, refresh
- [ ] Notification dispatch: single, multiple, chat
- [ ] All notification types: message, group, topic, signature, lock, invite
- [ ] Foreground: in-app toast
- [ ] Background: system notification
- [ ] Deep linking: all routes
- [ ] Muted chat: skip
- [ ] Badge count
- [ ] `go test ./...` pass

### Review Checklist:
- [ ] Notification content in Indonesian
- [ ] Deep link routes cover all screens
- [ ] FCM credentials configured
- [ ] iOS APNs support via FCM
- [ ] Privacy: no sensitive content in notification body
- [ ] Commit: `feat(notif): implement push notification system`
