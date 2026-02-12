# Phase 20: Cloud Backup

> Implementasi backup dan restore data pengguna ke cloud storage.
> Google Drive (Android) dan iCloud (iOS).

**Estimasi:** 3 hari
**Dependency:** Phase 19 (Local Storage)
**Output:** Backup/restore flow untuk chat history dan media.

---

## Task 20.1: Backup Service (Backend)

**Input:** Phase 19 local data, Phase 11 media storage
**Output:** Backend support untuk backup export/import

### Steps:
1. Buat `internal/service/backup_service.go`:
   ```go
   type BackupService interface {
       ExportUserData(ctx context.Context, userID uuid.UUID) (*BackupBundle, error)
       ImportUserData(ctx context.Context, userID uuid.UUID, bundle *BackupBundle) error
       GetBackupHistory(ctx context.Context, userID uuid.UUID) ([]*BackupRecord, error)
   }

   type BackupBundle struct {
       Version   int       `json:"version"` // schema version for migration
       UserID    string    `json:"userId"`
       CreatedAt time.Time `json:"createdAt"`
       Data      BackupData `json:"data"`
   }

   type BackupData struct {
       Profile    *UserProfile     `json:"profile"`
       Chats      []*ChatExport    `json:"chats"`
       Messages   []*MessageExport `json:"messages"`
       Contacts   []*ContactExport `json:"contacts"`
       Documents  []*DocumentExport `json:"documents"`
       Entities   []*EntityExport  `json:"entities"`
   }

   type BackupRecord struct {
       ID        uuid.UUID `json:"id"`
       UserID    uuid.UUID `json:"userId"`
       SizeBytes int64     `json:"sizeBytes"`
       Platform  string    `json:"platform"` // google_drive, icloud
       Status    string    `json:"status"`   // completed, failed
       CreatedAt time.Time `json:"createdAt"`
   }
   ```
2. Export endpoint:
   ```go
   // POST /api/v1/backup/export
   // Returns backup bundle as JSON (paginated for large datasets)
   // - Messages: last 10,000 per chat (or configurable)
   // - Media: URLs only (media files backed up separately)
   // - Documents: metadata + blocks (not media)

   GET /api/v1/backup/export/chats?offset=0&limit=100
   GET /api/v1/backup/export/messages?chatId=xxx&offset=0&limit=1000
   GET /api/v1/backup/export/documents?offset=0&limit=100
   ```
3. Import endpoint:
   ```go
   // POST /api/v1/backup/import
   // Accepts backup bundle
   // Merge strategy: skip existing (by server_id), insert new
   ```
4. Backup record tracking:
   ```sql
   CREATE TABLE backup_records (
       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
       user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
       size_bytes BIGINT NOT NULL,
       platform VARCHAR(20) NOT NULL,
       status VARCHAR(20) NOT NULL DEFAULT 'completed',
       created_at TIMESTAMPTZ DEFAULT NOW()
   );
   ```

### Acceptance Criteria:
- [ ] Export: user data bundled as JSON
- [ ] Import: merge with skip-existing strategy
- [ ] Backup history tracked
- [ ] Paginated export for large datasets
- [ ] Media: URLs only (not raw files)

### Testing:
- [ ] Unit test: export produces valid bundle
- [ ] Unit test: import merges correctly
- [ ] Unit test: skip existing on import
- [ ] Integration test: export → import roundtrip

---

## Task 20.2: Google Drive Backup (Android)

**Input:** Task 20.1
**Output:** Backup/restore via Google Drive API

### Steps:
1. Setup Google Drive API:
   ```typescript
   // src/services/GoogleDriveBackup.ts
   import { GoogleSignin } from '@react-native-google-signin/google-signin';
   import { GDrive, MimeTypes } from '@robinbobin/react-native-google-drive-api-wrapper';

   const BACKUP_FOLDER = 'Chatat Backup';
   const BACKUP_FILE = 'chatat-backup.json';

   class GoogleDriveBackupService {
     private gdrive: GDrive;

     async initialize(): Promise<void> {
       await GoogleSignin.configure({
         scopes: ['https://www.googleapis.com/auth/drive.file'],
       });
     }

     async signIn(): Promise<void> {
       await GoogleSignin.hasPlayServices();
       const userInfo = await GoogleSignin.signIn();
       const tokens = await GoogleSignin.getTokens();

       this.gdrive = new GDrive();
       this.gdrive.accessToken = tokens.accessToken;
     }

     async backup(): Promise<void> {
       // 1. Get backup data from server (paginated)
       const bundle = await this.collectBackupData();

       // 2. Get or create backup folder
       const folderId = await this.getOrCreateFolder(BACKUP_FOLDER);

       // 3. Upload backup as JSON file
       const backupJson = JSON.stringify(bundle);
       const fileName = `chatat-${new Date().toISOString().split('T')[0]}.json`;

       await this.gdrive.files
         .newMultipartUploader()
         .setData(backupJson, MimeTypes.JSON)
         .setRequestBody({
           name: fileName,
           parents: [folderId],
         })
         .execute();

       // 4. Log backup record
       await api.logBackup({
         sizeBytes: new Blob([backupJson]).size,
         platform: 'google_drive',
       });
     }

     async restore(): Promise<BackupBundle | null> {
       // 1. Find backup folder
       const folderId = await this.findFolder(BACKUP_FOLDER);
       if (!folderId) return null;

       // 2. List backup files (sorted by date desc)
       const files = await this.gdrive.files.list({
         q: `'${folderId}' in parents`,
         orderBy: 'createdTime desc',
         pageSize: 10,
       });

       if (!files.files?.length) return null;

       // 3. Download latest
       const latestFile = files.files[0];
       const content = await this.gdrive.files.getContent(latestFile.id);

       return JSON.parse(content) as BackupBundle;
     }

     async listBackups(): Promise<BackupFile[]> {
       const folderId = await this.findFolder(BACKUP_FOLDER);
       if (!folderId) return [];

       const files = await this.gdrive.files.list({
         q: `'${folderId}' in parents`,
         orderBy: 'createdTime desc',
         fields: 'files(id, name, size, createdTime)',
       });

       return files.files?.map((f) => ({
         id: f.id,
         name: f.name,
         sizeBytes: parseInt(f.size || '0'),
         createdAt: f.createdTime,
       })) || [];
     }
   }
   ```
2. Backup progress tracking:
   - Show progress percentage during backup/restore
   - Estimate time remaining
   - Cancel option

### Acceptance Criteria:
- [ ] Google Sign-In for Drive access
- [ ] Create backup folder in Drive
- [ ] Upload backup JSON
- [ ] List available backups
- [ ] Restore from selected backup
- [ ] Backup record logged
- [ ] Progress indicator

### Testing:
- [ ] Unit test: backup data collection
- [ ] Unit test: folder creation
- [ ] Integration test: backup + list + restore (mock Drive API)

---

## Task 20.3: iCloud Backup (iOS)

**Input:** Task 20.1
**Output:** Backup/restore via iCloud

### Steps:
1. Setup iCloud storage:
   ```typescript
   // src/services/iCloudBackup.ts
   import CloudStore from 'react-native-cloud-store';

   const BACKUP_DIR = 'Documents/Chatat';

   class ICloudBackupService {
     async isAvailable(): Promise<boolean> {
       return await CloudStore.isICloudAvailable();
     }

     async backup(): Promise<void> {
       // 1. Collect backup data
       const bundle = await this.collectBackupData();
       const backupJson = JSON.stringify(bundle);
       const fileName = `chatat-${new Date().toISOString().split('T')[0]}.json`;

       // 2. Write to iCloud
       const path = `${BACKUP_DIR}/${fileName}`;
       await CloudStore.writeFile(path, backupJson, { override: true });

       // 3. Log backup record
       await api.logBackup({
         sizeBytes: backupJson.length,
         platform: 'icloud',
       });
     }

     async restore(): Promise<BackupBundle | null> {
       // 1. List backup files
       const backups = await this.listBackups();
       if (backups.length === 0) return null;

       // 2. Read latest
       const latest = backups[0];
       const content = await CloudStore.readFile(latest.path);

       return JSON.parse(content) as BackupBundle;
     }

     async listBackups(): Promise<BackupFile[]> {
       try {
         const exists = await CloudStore.exist(BACKUP_DIR);
         if (!exists) return [];

         const files = await CloudStore.readDir(BACKUP_DIR);
         return files
           .filter((f) => f.endsWith('.json'))
           .sort()
           .reverse()
           .map((f) => ({
             path: `${BACKUP_DIR}/${f}`,
             name: f,
             createdAt: this.parseDate(f),
           }));
       } catch {
         return [];
       }
     }
   }
   ```
2. iCloud entitlement:
   - Add iCloud capability in Xcode
   - Enable iCloud Documents
   - Set container identifier

### Acceptance Criteria:
- [ ] iCloud availability check
- [ ] Write backup to iCloud Documents
- [ ] List available backups
- [ ] Restore from selected backup
- [ ] Handle iCloud not available gracefully
- [ ] Progress indicator

### Testing:
- [ ] Unit test: backup data collection
- [ ] Integration test: file write/read (mock CloudStore)

---

## Task 20.4: Backup UI

**Input:** Task 20.2, 20.3
**Output:** Backup management screen

### Steps:
1. Backup settings screen:
   ```typescript
   // src/screens/BackupScreen.tsx
   const BackupScreen: React.FC = () => {
     const platform = Platform.OS === 'ios' ? 'icloud' : 'google_drive';

     return (
       <ScrollView style={styles.container}>
         <SectionHeader title="Cadangan Data" />

         {/* Backup status */}
         <InfoCard>
           <Text style={styles.label}>Cadangan terakhir</Text>
           <Text style={styles.value}>
             {lastBackup ? formatDate(lastBackup.createdAt) : 'Belum pernah'}
           </Text>
           <Text style={styles.label}>Ukuran</Text>
           <Text style={styles.value}>
             {lastBackup ? formatBytes(lastBackup.sizeBytes) : '-'}
           </Text>
         </InfoCard>

         {/* Backup now */}
         <Button
           title="Cadangkan Sekarang"
           onPress={handleBackup}
           loading={isBackingUp}
           icon="cloud-upload"
         />

         {/* Backup options */}
         <SettingRow
           label="Cadangkan Otomatis"
           type="switch"
           value={autoBackup}
           onToggle={setAutoBackup}
         />
         <SettingRow
           label="Frekuensi"
           type="select"
           value={backupFrequency}
           options={[
             { label: 'Harian', value: 'daily' },
             { label: 'Mingguan', value: 'weekly' },
             { label: 'Bulanan', value: 'monthly' },
           ]}
           onSelect={setBackupFrequency}
         />

         {/* Available backups */}
         <SectionHeader title="Cadangan Tersedia" />
         <FlatList
           data={backups}
           renderItem={({ item }) => (
             <BackupListItem
               backup={item}
               onRestore={() => handleRestore(item)}
             />
           )}
           ListEmptyComponent={<Text>Tidak ada cadangan</Text>}
         />

         {/* Platform info */}
         <InfoCard style={styles.platformInfo}>
           {platform === 'icloud' ? (
             <Text>Data dicadangkan ke iCloud</Text>
           ) : (
             <Text>Data dicadangkan ke Google Drive</Text>
           )}
         </InfoCard>
       </ScrollView>
     );
   };
   ```
2. Backup progress modal:
   ```typescript
   const BackupProgressModal: React.FC<Props> = ({ visible, progress, status }) => (
     <Modal visible={visible} transparent>
       <View style={styles.overlay}>
         <View style={styles.modal}>
           <Text style={styles.title}>
             {status === 'backing_up' ? 'Mencadangkan...' : 'Memulihkan...'}
           </Text>
           <ProgressBar progress={progress} />
           <Text style={styles.percent}>{Math.round(progress * 100)}%</Text>
           <Button title="Batal" variant="text" onPress={onCancel} />
         </View>
       </View>
     </Modal>
   );
   ```
3. Restore confirmation:
   - Warning: "Data saat ini akan digabungkan dengan cadangan"
   - Confirm button

### Acceptance Criteria:
- [ ] Backup screen shows last backup info
- [ ] "Cadangkan Sekarang" button
- [ ] Auto backup toggle + frequency
- [ ] List available backups
- [ ] Restore from backup with confirmation
- [ ] Progress indicator during backup/restore
- [ ] Platform-specific: iCloud (iOS) / Google Drive (Android)

### Testing:
- [ ] Component test: BackupScreen renders
- [ ] Component test: progress modal
- [ ] Component test: restore confirmation
- [ ] Component test: platform detection

---

## Phase 20 Review

### Testing Checklist:
- [ ] Backend: export/import API
- [ ] Google Drive: backup, list, restore
- [ ] iCloud: backup, list, restore
- [ ] Backup UI: status, trigger, list, restore
- [ ] Progress indicator during operations
- [ ] Backup records tracked
- [ ] Platform detection (iOS → iCloud, Android → Drive)

### Review Checklist:
- [ ] Backup sesuai `spesifikasi-chatat.md` section 9.1
- [ ] Google Drive sesuai spec 9.1.1
- [ ] iCloud sesuai spec 9.1.2
- [ ] Indonesian labels
- [ ] Error handling: no iCloud, no Drive access
- [ ] Commit: `feat(backup): implement cloud backup (Google Drive + iCloud)`
