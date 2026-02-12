# Phase 06: Mobile App Shell

> Setup navigasi, theme system, design system, dan shared components.
> Phase ini menghasilkan app shell siap pakai untuk semua screen.

**Estimasi:** 4 hari
**Dependency:** Phase 01 (Project Setup)
**Output:** App navigasi berfungsi, theme WA-style, komponen dasar siap.

---

## Task 6.1: Navigation Structure

**Input:** React Native project dari Phase 01
**Output:** Full navigation tree dengan React Navigation

### Steps:
1. Setup navigation di `src/navigation/`:
   ```
   navigation/
   ├── RootNavigator.tsx
   ├── AuthNavigator.tsx
   ├── MainNavigator.tsx
   ├── ChatStackNavigator.tsx
   ├── DocumentStackNavigator.tsx
   └── types.ts
   ```
2. Buat `RootNavigator.tsx`:
   ```tsx
   // Conditional: AuthNavigator or MainNavigator based on auth state
   const RootNavigator = () => {
     const isAuthenticated = useAuthStore(state => state.isAuthenticated);
     return isAuthenticated ? <MainNavigator /> : <AuthNavigator />;
   };
   ```
3. Buat `AuthNavigator.tsx`:
   - Screen: PhoneInput → OTPVerify → ProfileSetup
4. Buat `MainNavigator.tsx`:
   - Bottom Tab Navigator dengan 2 tab:
     - **Chat** (💬): ChatListScreen
     - **Dokumen** (📄): DocumentListScreen
5. Buat `ChatStackNavigator.tsx`:
   - ChatList → ChatScreen → ChatInfo
   - ChatList → ContactList → ChatScreen
   - ChatList → CreateGroup → ChatScreen
   - ChatScreen → TopicList → TopicScreen
   - ChatScreen → DocumentEditor
6. Buat `DocumentStackNavigator.tsx`:
   - DocumentList → DocumentEditor
   - DocumentList → DocumentViewer (locked docs)
7. Navigation types di `types.ts`:
   ```tsx
   type RootStackParamList = {
     Auth: undefined;
     Main: undefined;
   };

   type AuthStackParamList = {
     PhoneInput: undefined;
     OTPVerify: { phone: string; method: 'sms' | 'reverse' };
     ReverseOTPWait: { sessionId: string; waNumber: string; code: string };
     ProfileSetup: undefined;
   };

   type MainTabParamList = {
     ChatTab: undefined;
     DocumentTab: undefined;
   };

   type ChatStackParamList = {
     ChatList: undefined;
     Chat: { chatId: string; chatType: 'personal' | 'group' };
     ChatInfo: { chatId: string };
     ContactList: undefined;
     CreateGroup: undefined;
     TopicList: { chatId: string };
     Topic: { topicId: string };
     DocumentEditor: { documentId?: string; contextType?: string; contextId?: string };
   };
   ```

### Acceptance Criteria:
- [x] Auth flow: Phone → OTP → Profile → Main
- [x] Main tabs: Chat dan Dokumen
- [x] Chat stack: list → chat → info/topics/documents
- [x] Document stack: list → editor/viewer
- [x] Navigation types fully typed
- [x] Deep linking preparation (params defined)

### Testing:
- [x] Navigation test: auth flow complete
- [x] Navigation test: tab switching
- [x] Navigation test: stack push/pop
- [x] Navigation test: deep params passing

---

## Task 6.2: Theme System (WhatsApp Dark)

**Input:** Color palette dari `spesifikasi-chatat.md` section 9
**Output:** Theme provider dengan WA-style dark colors

### Steps:
1. Buat `src/theme/colors.ts`:
   ```tsx
   export const colors = {
     // Background
     background: '#0F1117',
     surface: '#1A1D27',
     surface2: '#222637',
     border: '#2E3348',

     // Text
     textPrimary: '#E8EAF0',
     textMuted: '#6B7280',

     // Accent
     green: '#6EE7B7',
     purple: '#818CF8',
     blue: '#60A5FA',
     red: '#F87171',
     yellow: '#FBBF24',

     // Chat bubbles
     bubbleSelf: '#1B3A2D',     // dark green for own messages
     bubbleOther: '#222637',    // surface2 for others
     bubbleSelfText: '#E8EAF0',
     bubbleOtherText: '#E8EAF0',

     // Status
     online: '#6EE7B7',
     offline: '#6B7280',

     // Misc
     overlay: 'rgba(0, 0, 0, 0.5)',
     inputBackground: '#1A1D27',
     tabBarBackground: '#1A1D27',
     headerBackground: '#1A1D27',
   };
   ```
2. Buat `src/theme/typography.ts`:
   ```tsx
   export const typography = {
     // Font families
     fontFamily: {
       ui: 'PlusJakartaSans',
       document: 'Inter',
       code: 'JetBrainsMono',
     },
     // Font sizes
     fontSize: {
       xs: 11,
       sm: 13,
       md: 15,
       lg: 17,
       xl: 20,
       xxl: 24,
       h1: 28,
       h2: 24,
       h3: 20,
     },
     // Line heights
     lineHeight: {
       tight: 1.2,
       normal: 1.5,
       relaxed: 1.75,
     },
   };
   ```
3. Buat `src/theme/spacing.ts`:
   ```tsx
   export const spacing = {
     xs: 4,
     sm: 8,
     md: 12,
     lg: 16,
     xl: 20,
     xxl: 24,
     xxxl: 32,
   };
   ```
4. Buat `src/theme/index.ts` yang export semua
5. Install dan configure custom fonts:
   - Plus Jakarta Sans (UI)
   - Inter (dokumen)
   - JetBrains Mono (code blocks)
   - Link fonts via `react-native.config.js` atau Expo config
6. Setup `StatusBar` style: light-content (untuk dark background)

### Acceptance Criteria:
- [x] Semua warna sesuai spec section 9.2
- [x] Typography sesuai spec section 9.3
- [x] Custom fonts loaded dan berfungsi
- [x] Consistent spacing scale
- [x] StatusBar light content untuk dark theme

### Testing:
- [x] Visual test: colors match spec
- [x] Visual test: fonts render correctly
- [x] Visual test: spacing consistent

---

## Task 6.3: Shared UI Components

**Input:** Task 6.2 (theme)
**Output:** Reusable UI components

### Steps:
1. Buat `src/components/ui/`:
   - **Avatar.tsx**: Emoji avatar dengan background bulat
     ```tsx
     type AvatarProps = {
       emoji: string;
       size?: 'sm' | 'md' | 'lg';
       online?: boolean;  // green dot indicator
     };
     ```
   - **Badge.tsx**: Unread count badge, status badge
     ```tsx
     type BadgeProps = {
       count?: number;
       variant: 'unread' | 'draft' | 'locked' | 'signature';
     };
     ```
   - **Button.tsx**: Primary (green), secondary, danger, ghost
   - **TextInput.tsx**: Themed input field
   - **IconButton.tsx**: Icon-only button
   - **Divider.tsx**: Horizontal separator
   - **BottomSheet.tsx**: Modal dari bawah layar
   - **Pressable.tsx**: Themed pressable with haptic feedback

2. Buat `src/components/shared/`:
   - **LoadingScreen.tsx**: Full-screen loading spinner
   - **EmptyState.tsx**: Illustration + message untuk empty lists
   - **ErrorState.tsx**: Error message + retry button
   - **SearchBar.tsx**: Search input dengan ikon
   - **StatusText.tsx**: "terakhir dilihat pukul HH:MM" / "online"
   - **DateSeparator.tsx**: Date divider untuk chat (Hari ini, Kemarin, tanggal)
   - **ConfirmDialog.tsx**: Confirmation modal

3. Buat `src/components/layout/`:
   - **Header.tsx**: App header dengan logo + search + profile
   - **TabBar.tsx**: Custom bottom tab bar
   - **FAB.tsx**: Floating Action Button (hijau, pojok kanan bawah)
   - **ScreenContainer.tsx**: SafeAreaView wrapper

### Acceptance Criteria:
- [x] Avatar: emoji render, online indicator
- [x] Badge: count > 99 shows "99+"
- [x] Button: 4 variants, disabled state, loading state
- [ ] BottomSheet: slide up, backdrop dismiss (deferred to Phase 07)
- [x] All components use theme colors/typography
- [x] FAB: positioned fixed, shadow, press animation

### Testing:
- [x] Component test: Avatar renders with emoji
- [x] Component test: Badge displays count
- [x] Component test: Button variants and states
- [x] Component test: SearchBar input handling
- [ ] Snapshot tests for all shared components (deferred to Phase 24)

---

## Task 6.4: Auth Screens

**Input:** Task 6.1 (navigation), 6.3 (components)
**Output:** Auth flow screens (UI only, backend integration di Phase 04)

### Steps:
1. Buat `src/screens/auth/PhoneInputScreen.tsx`:
   - Logo Chatat di atas
   - Country code picker (+62 default)
   - Phone number input
   - "Lanjut dengan SMS OTP" button
   - "Lanjut dengan WhatsApp" button (reverse OTP)
   - Teks info: privacy policy reference
2. Buat `src/screens/auth/OTPVerifyScreen.tsx`:
   - Teks: "Masukkan kode 6 digit yang dikirim ke +62xxx"
   - 6-digit input (individual boxes, auto-focus next)
   - Timer countdown (resend after 60s)
   - "Kirim Ulang" button (disabled until timer)
   - Auto-submit saat 6 digit terisi
3. Buat `src/screens/auth/ReverseOTPWaitScreen.tsx`:
   - Teks: "Kirim pesan berikut ke WhatsApp"
   - Nomor WA server (large, copyable)
   - Kode unik (large, bold)
   - "Buka WhatsApp" button (deep link ke WA)
   - Polling indicator: "Menunggu verifikasi..."
   - Timer countdown (expire after 5 min)
4. Buat `src/screens/auth/ProfileSetupScreen.tsx`:
   - Avatar picker (emoji grid)
   - Name input (required)
   - "Mulai" button
   - Skip not allowed (name must be set)

### Acceptance Criteria:
- [x] Phone input: country code + phone number
- [x] OTP input: 6 digit, auto-focus, auto-submit
- [x] Reverse OTP: WA number + code displayed, open WA button
- [x] Profile setup: emoji picker + name input
- [x] All screens follow dark theme
- [x] Input validation (empty, too short)
- [x] Loading states on submit buttons

### Testing:
- [x] Component test: phone input validation
- [x] Component test: OTP auto-focus behavior
- [x] Component test: profile setup validation
- [ ] Snapshot tests for auth screens (deferred to Phase 24)

---

## Task 6.5: API Service Layer (Mobile)

**Input:** Task 6.1 selesai
**Output:** API client untuk komunikasi dengan backend

### Steps:
1. Buat `src/services/api/client.ts`:
   ```tsx
   import axios from 'axios';

   const apiClient = axios.create({
     baseURL: Config.API_URL,
     timeout: 30000,
     headers: { 'Content-Type': 'application/json' },
   });

   // Request interceptor: add auth token
   apiClient.interceptors.request.use(config => {
     const token = useAuthStore.getState().accessToken;
     if (token) config.headers.Authorization = `Bearer ${token}`;
     return config;
   });

   // Response interceptor: handle 401, refresh token
   apiClient.interceptors.response.use(
     response => response,
     async error => {
       if (error.response?.status === 401) {
         // Try refresh token
         // If refresh fails → logout
       }
       return Promise.reject(error);
     }
   );
   ```
2. Buat `src/services/api/auth.ts`:
   ```tsx
   export const authApi = {
     sendOTP: (phone: string) => apiClient.post('/auth/otp/send', { phone }),
     verifyOTP: (phone: string, code: string) => apiClient.post('/auth/otp/verify', { phone, code }),
     initReverseOTP: (phone: string) => apiClient.post('/auth/reverse-otp/init', { phone }),
     checkReverseOTP: (sessionId: string) => apiClient.post('/auth/reverse-otp/check', { sessionId }),
     refreshToken: (token: string) => apiClient.post('/auth/refresh', { refreshToken: token }),
     logout: () => apiClient.post('/auth/logout'),
   };
   ```
3. Buat API service files untuk setiap domain (placeholder):
   - `src/services/api/users.ts`
   - `src/services/api/contacts.ts`
   - `src/services/api/chats.ts`
   - `src/services/api/topics.ts`
   - `src/services/api/documents.ts`
   - `src/services/api/entities.ts`
   - `src/services/api/media.ts`
4. Buat `src/stores/authStore.ts` (Zustand):
   ```tsx
   interface AuthState {
     isAuthenticated: boolean;
     accessToken: string | null;
     refreshToken: string | null;
     user: User | null;
     isNewUser: boolean;
     login: (tokens: TokenPair, user: User, isNew: boolean) => void;
     logout: () => void;
     setUser: (user: User) => void;
   }
   ```

### Acceptance Criteria:
- [x] API client dengan base URL configurable
- [x] Auth token auto-attached to requests
- [x] 401 → auto refresh → retry request
- [x] Refresh fail → logout
- [x] Auth store persisted (AsyncStorage)
- [x] All API service files created (placeholder)
- [x] Type safety: all API responses typed

### Testing:
- [x] Unit test: API client interceptors
- [x] Unit test: token refresh flow
- [x] Unit test: auth store actions

---

## Phase 06 Review

### Testing Checklist:
- [x] App launches dengan splash → auth/main
- [x] Auth flow: phone → OTP → profile → main screen
- [x] Tab bar: Chat dan Dokumen tabs switch
- [x] Theme: dark background, green accents, correct fonts
- [x] Components: all shared components render correctly
- [x] Navigation: push/pop/tab switch smooth
- [x] API client: configured with interceptors

### Review Checklist:
- [x] Navigation sesuai `spesifikasi-chatat.md` section 7
- [x] Colors sesuai `spesifikasi-chatat.md` section 9.2
- [x] Typography sesuai `spesifikasi-chatat.md` section 9.3
- [x] Component naming sesuai `docs/naming-conventions.md`
- [x] Style sesuai `docs/react-native-style-guide.md`
- [x] Commit: `feat(mobile): implement app shell and navigation`
