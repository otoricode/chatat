# Phase 21: Settings & Preferences

> Implementasi halaman pengaturan, edit profil, preferensi notifikasi,
> manajemen penyimpanan, dan fitur logout.

**Estimasi:** 2 hari
**Dependency:** Phase 06 (Mobile Shell), Phase 18 (i18n), Phase 20 (Backup)
**Output:** Settings screen lengkap dengan semua preferensi.

---

## Task 21.1: Settings Screen

**Input:** Phase 06 design system
**Output:** Main settings screen

### Steps:
1. Buat `src/screens/SettingsScreen.tsx`:
   ```typescript
   const SettingsScreen: React.FC = () => {
     const { t } = useTranslation();
     const user = useAuthStore((s) => s.user);

     return (
       <ScrollView style={styles.container}>
         {/* Profile header */}
         <TouchableOpacity
           style={styles.profileHeader}
           onPress={() => navigation.navigate('EditProfile')}
         >
           <Avatar name={user.name} url={user.avatarUrl} size={64} />
           <View style={styles.profileInfo}>
             <Text style={styles.name}>{user.name}</Text>
             <Text style={styles.phone}>{user.phone}</Text>
             <Text style={styles.status}>{user.statusText || t('settings.noStatus')}</Text>
           </View>
           <ChevronRightIcon size={20} color="#6B7280" />
         </TouchableOpacity>

         {/* Sections */}
         <SettingSection title={t('settings.account')}>
           <SettingRow
             icon="globe"
             label={t('settings.language')}
             value={getCurrentLanguageLabel()}
             onPress={() => navigation.navigate('LanguageSettings')}
           />
           <SettingRow
             icon="bell"
             label={t('settings.notifications')}
             onPress={() => navigation.navigate('NotificationSettings')}
           />
           <SettingRow
             icon="cloud"
             label={t('settings.backup')}
             onPress={() => navigation.navigate('BackupScreen')}
           />
         </SettingSection>

         <SettingSection title={t('settings.storage')}>
           <SettingRow
             icon="hard-drive"
             label={t('settings.storageUsage')}
             value={formatBytes(storageUsage)}
             onPress={() => navigation.navigate('StorageSettings')}
           />
           <SettingRow
             icon="download"
             label={t('settings.autoDownload')}
             onPress={() => navigation.navigate('AutoDownloadSettings')}
           />
         </SettingSection>

         <SettingSection title={t('settings.about')}>
           <SettingRow
             icon="info"
             label={t('settings.about')}
             value="v1.0.0"
             onPress={() => navigation.navigate('AboutScreen')}
           />
           <SettingRow
             icon="shield"
             label={t('settings.privacy')}
             onPress={() => navigation.navigate('PrivacySettings')}
           />
         </SettingSection>

         <TouchableOpacity style={styles.logoutButton} onPress={handleLogout}>
           <LogOutIcon size={20} color="#EF4444" />
           <Text style={styles.logoutText}>{t('settings.logout')}</Text>
         </TouchableOpacity>

         <Text style={styles.version}>Chatat v1.0.0</Text>
       </ScrollView>
     );
   };
   ```

### Acceptance Criteria:
- [ ] Profile header: avatar, name, phone, status
- [ ] Tap profile → edit profile screen
- [ ] All setting sections rendered
- [ ] Logout button with red styling
- [ ] Version number shown
- [ ] Navigation to sub-screens

### Testing:
- [ ] Component test: SettingsScreen renders all sections
- [ ] Component test: profile header renders user info
- [ ] Component test: logout button renders

---

## Task 21.2: Edit Profile Screen

**Input:** Phase 05 (User/Contact)
**Output:** Profile editing with avatar upload

### Steps:
1. Buat `src/screens/EditProfileScreen.tsx`:
   ```typescript
   const EditProfileScreen: React.FC = () => {
     const { t } = useTranslation();
     const user = useAuthStore((s) => s.user);
     const [name, setName] = useState(user.name);
     const [statusText, setStatusText] = useState(user.statusText || '');
     const [avatar, setAvatar] = useState<string | null>(user.avatarUrl);

     const handlePickAvatar = async () => {
       const result = await ImagePicker.launchImageLibrary({
         mediaType: 'photo',
         maxWidth: 512,
         maxHeight: 512,
         quality: 0.8,
       });

       if (result.assets?.[0]) {
         const uploaded = await api.uploadAvatar(result.assets[0]);
         setAvatar(uploaded.url);
       }
     };

     const handleSave = async () => {
       await api.updateProfile({ name, statusText, avatarUrl: avatar });
       useAuthStore.getState().updateUser({ name, statusText, avatarUrl: avatar });
       navigation.goBack();
     };

     return (
       <KeyboardAvoidingView style={styles.container}>
         <TouchableOpacity style={styles.avatarContainer} onPress={handlePickAvatar}>
           <Avatar name={name} url={avatar} size={96} />
           <View style={styles.cameraIcon}>
             <CameraIcon size={16} color="#FFFFFF" />
           </View>
         </TouchableOpacity>

         <TextInput
           label={t('auth.nameLabel')}
           value={name}
           onChangeText={setName}
           maxLength={50}
         />
         <Text style={styles.charCount}>{name.length}/50</Text>

         <TextInput
           label={t('settings.statusLabel')}
           value={statusText}
           onChangeText={setStatusText}
           maxLength={140}
           placeholder={t('settings.statusPlaceholder')}
         />
         <Text style={styles.charCount}>{statusText.length}/140</Text>

         <View style={styles.phoneSection}>
           <Text style={styles.label}>{t('auth.phoneLabel')}</Text>
           <Text style={styles.phone}>{user.phone}</Text>
           <Text style={styles.hint}>{t('settings.phoneCannotChange')}</Text>
         </View>

         <Button title={t('common.save')} onPress={handleSave} />
       </KeyboardAvoidingView>
     );
   };
   ```

### Acceptance Criteria:
- [ ] Edit name (max 50 chars)
- [ ] Edit status text (max 140 chars)
- [ ] Upload avatar from gallery
- [ ] Phone number shown (not editable)
- [ ] Save updates profile
- [ ] Character count shown

### Testing:
- [ ] Component test: form renders with current data
- [ ] Component test: avatar picker
- [ ] Component test: character count
- [ ] Component test: save button

---

## Task 21.3: Notification & Storage Preferences

**Input:** Phase 16 (Push Notifications)
**Output:** Notification and storage settings screens

### Steps:
1. Notification settings:
   ```typescript
   const NotificationSettingsScreen: React.FC = () => {
     return (
       <ScrollView style={styles.container}>
         <SettingSection title={t('settings.messageNotifications')}>
           <SettingRow
             label={t('settings.showPreview')}
             type="switch"
             value={showPreview}
             onToggle={setShowPreview}
           />
           <SettingRow
             label={t('settings.sound')}
             type="switch"
             value={soundEnabled}
             onToggle={setSoundEnabled}
           />
           <SettingRow
             label={t('settings.vibration')}
             type="switch"
             value={vibrationEnabled}
             onToggle={setVibrationEnabled}
           />
         </SettingSection>

         <SettingSection title={t('settings.groupNotifications')}>
           <SettingRow
             label={t('settings.groupAlerts')}
             type="switch"
             value={groupAlerts}
             onToggle={setGroupAlerts}
           />
         </SettingSection>
       </ScrollView>
     );
   };
   ```
2. Storage settings:
   ```typescript
   const StorageSettingsScreen: React.FC = () => {
     return (
       <ScrollView style={styles.container}>
         {/* Storage usage breakdown */}
         <StorageChart data={storageData} />

         <SettingSection title={t('settings.storageBreakdown')}>
           <StorageRow label={t('settings.messages')} size={storageData.messages} />
           <StorageRow label={t('settings.media')} size={storageData.media} />
           <StorageRow label={t('settings.documents')} size={storageData.documents} />
           <StorageRow label={t('settings.cache')} size={storageData.cache} />
         </SettingSection>

         <Button
           title={t('settings.clearCache')}
           variant="outline"
           onPress={handleClearCache}
         />
       </ScrollView>
     );
   };
   ```
3. Auto-download settings:
   - When on WiFi: auto-download all
   - When on cellular: auto-download < 5MB
   - When on cellular: ask for > 5MB

### Acceptance Criteria:
- [ ] Notification toggles: preview, sound, vibration
- [ ] Group notification settings
- [ ] Storage breakdown with chart
- [ ] Clear cache button
- [ ] Auto-download preferences
- [ ] Settings persisted to AsyncStorage

### Testing:
- [ ] Component test: notification settings toggles
- [ ] Component test: storage breakdown
- [ ] Component test: auto-download settings

---

## Task 21.4: Logout & Language Switch

**Input:** Phase 04 (Auth), Phase 18 (i18n)
**Output:** Logout flow dan language switch

### Steps:
1. Logout flow:
   ```typescript
   const handleLogout = async () => {
     Alert.alert(
       t('settings.logoutConfirm'),
       t('settings.logoutMessage'),
       [
         { text: t('common.cancel'), style: 'cancel' },
         {
           text: t('settings.logout'),
           style: 'destructive',
           onPress: async () => {
             // 1. Unregister push token
             await api.unregisterDevice();
             // 2. Clear tokens
             await SecureStore.deleteItemAsync('accessToken');
             await SecureStore.deleteItemAsync('refreshToken');
             // 3. Clear stores
             useAuthStore.getState().reset();
             useChatStore.getState().reset();
             // 4. Optionally clear local DB
             // await database.write(async () => { await database.unsafeResetDatabase(); });
             // 5. Navigate to auth
             navigation.reset({ index: 0, routes: [{ name: 'Auth' }] });
           },
         },
       ]
     );
   };
   ```
2. Language switch screen:
   ```typescript
   const LanguageSettingsScreen: React.FC = () => {
     const { i18n, t } = useTranslation();

     const languages = [
       { code: 'id', label: 'Bahasa Indonesia', flag: '🇮🇩' },
       { code: 'en', label: 'English', flag: '🇬🇧' },
       { code: 'ar', label: 'العربية', flag: '🇸🇦' },
     ];

     const handleSelect = async (code: string) => {
       await setLanguage(code as 'id' | 'en' | 'ar');
       await api.updateProfile({ language: code });
       // RTL change for Arabic requires restart
       if (code === 'ar' || i18n.language === 'ar') {
         applyRTL(code);
       }
     };

     return (
       <FlatList
         data={languages}
         renderItem={({ item }) => (
           <TouchableOpacity
             style={[styles.langRow, i18n.language === item.code && styles.active]}
             onPress={() => handleSelect(item.code)}
           >
             <Text style={styles.flag}>{item.flag}</Text>
             <Text style={styles.label}>{item.label}</Text>
             {i18n.language === item.code && (
               <CheckIcon size={20} color="#6EE7B7" />
             )}
           </TouchableOpacity>
         )}
       />
     );
   };
   ```

### Acceptance Criteria:
- [ ] Logout: confirmation dialog
- [ ] Logout: clears tokens, stores, push token
- [ ] Logout: navigates to auth screen
- [ ] Language switch: 3 options with flags
- [ ] Active language highlighted
- [ ] Arabic: triggers RTL + restart
- [ ] Language synced to server profile

### Testing:
- [ ] Unit test: logout clears all stores
- [ ] Component test: logout confirmation dialog
- [ ] Component test: language list renders
- [ ] Component test: active language check mark

---

## Phase 21 Review

### Testing Checklist:
- [ ] Settings screen: all sections render
- [ ] Edit profile: name, status, avatar
- [ ] Notification settings: toggles persist
- [ ] Storage settings: breakdown + clear cache
- [ ] Language switch: all 3 languages
- [ ] Logout: full cleanup flow
- [ ] Navigation between all settings screens

### Review Checklist:
- [ ] Settings sesuai `spesifikasi-chatat.md` section 9
- [ ] Indonesian labels (localized)
- [ ] Dark theme consistent
- [ ] Commit: `feat(settings): implement settings and preferences`
