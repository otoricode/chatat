# Phase 26: Beta Release

> Beta release ke TestFlight (iOS) dan Play Console (Android).
> Feedback collection, bug fixing, dan stabilization.

**Estimasi:** 4 hari
**Dependency:** Phase 25 (CI/CD), all feature phases
**Output:** Beta build distributed to testers.

---

## Task 26.1: App Store & Play Store Setup

**Input:** Chatat app
**Output:** Store listings dan certificates

### Steps:
1. Apple Developer setup:
   - App ID registration (com.otoritech.chatat)
   - Push notification certificate (APNs)
   - Provisioning profiles (development + distribution)
   - App Store Connect: create app
   - TestFlight: internal testing group
2. Google Play Console setup:
   - Create app (com.otoritech.chatat)
   - Upload signing key
   - Internal testing track
   - Play Console API access for CI
3. App metadata (both stores):
   ```
   App Name: Chatat
   Subtitle: Chat & Dokumen Kolaboratif
   Category: Social Networking / Productivity
   
   Short Description:
   "Chatat menggabungkan chat personal & grup dengan dokumen kolaboratif 
   bergaya Notion. Cocok untuk keluarga, komunitas, dan tim kecil."
   
   Keywords: chat, dokumen, kolaborasi, keluarga, grup, catatan, notulen
   
   Privacy Policy URL: https://chatat.app/privacy
   Support URL: https://chatat.app/support
   ```
4. App icons and screenshots:
   - App icon: 1024x1024 (iOS), 512x512 (Android)
   - Screenshots: iPhone 15 Pro Max, iPhone SE, iPad
   - Screenshots: Pixel 8, 10" tablet
   - Feature graphic (Android): 1024x500

### Acceptance Criteria:
- [ ] Apple Developer: App ID + certificates
- [ ] App Store Connect: app created
- [ ] TestFlight: testing group configured
- [ ] Google Play Console: app created
- [ ] Internal testing track configured
- [ ] App metadata filled
- [ ] Icons uploaded

### Testing:
- [ ] Certificates valid and working
- [ ] TestFlight upload succeeds
- [ ] Play Console upload succeeds

---

## Task 26.2: Beta Build & Distribution

**Input:** Task 26.1, Phase 25 CI/CD
**Output:** Beta builds distributed via TestFlight and Play Console

### Steps:
1. iOS build with Fastlane:
   ```ruby
   # mobile/ios/fastlane/Fastfile
   default_platform(:ios)

   platform :ios do
     desc "Build and upload to TestFlight"
     lane :beta do
       increment_build_number(
         build_number: ENV['BUILD_NUMBER'] || Time.now.to_i
       )

       build_app(
         workspace: "Chatat.xcworkspace",
         scheme: "Chatat",
         export_method: "app-store",
         output_directory: "./build",
         output_name: "Chatat.ipa"
       )

       upload_to_testflight(
         skip_waiting_for_build_processing: true,
         apple_id: ENV['APPLE_ID'],
       )
     end
   end
   ```
2. Android build with Fastlane:
   ```ruby
   # mobile/android/fastlane/Fastfile
   default_platform(:android)

   platform :android do
     desc "Build and upload to Play Console internal track"
     lane :beta do
       gradle(
         task: "bundle",
         build_type: "Release",
         project_dir: "."
       )

       upload_to_play_store(
         track: "internal",
         aab: "app/build/outputs/bundle/release/app-release.aab",
         skip_upload_metadata: true,
         skip_upload_images: true,
         skip_upload_screenshots: true,
       )
     end
   end
   ```
3. CI integration for beta:
   ```yaml
   # .github/workflows/beta.yml
   name: Beta Release
   on:
     push:
       tags: ['beta-*']
   jobs:
     ios-beta:
       runs-on: macos-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-node@v4
           with:
             node-version: '20'
         - run: cd mobile && yarn install --frozen-lockfile
         - run: cd mobile/ios && pod install
         - uses: ruby/setup-ruby@v1
           with:
             ruby-version: '3.2'
         - run: cd mobile/ios && bundle install
         - run: cd mobile/ios && bundle exec fastlane beta
           env:
             APPLE_ID: ${{ secrets.APPLE_ID }}
             MATCH_PASSWORD: ${{ secrets.MATCH_PASSWORD }}

     android-beta:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-java@v4
           with:
             distribution: 'temurin'
             java-version: '17'
         - uses: actions/setup-node@v4
           with:
             node-version: '20'
         - run: cd mobile && yarn install --frozen-lockfile
         - uses: ruby/setup-ruby@v1
           with:
             ruby-version: '3.2'
         - run: cd mobile/android && bundle install
         - run: cd mobile/android && bundle exec fastlane beta
           env:
             PLAY_STORE_JSON_KEY: ${{ secrets.PLAY_STORE_JSON_KEY }}
   ```

### Acceptance Criteria:
- [ ] iOS: build + TestFlight upload works
- [ ] Android: build + Play Console upload works
- [ ] Fastlane configured for both platforms
- [ ] CI workflow triggered by beta tag
- [ ] Build number auto-incremented

### Testing:
- [ ] Tag beta-v0.1.0 → both builds trigger
- [ ] TestFlight: build appears
- [ ] Play Console: build appears in internal track

---

## Task 26.3: Beta Testing & Feedback

**Input:** Task 26.2, testers group
**Output:** Structured feedback and bug tracking

### Steps:
1. Tester recruitment:
   - Internal team: 5-10 people
   - External beta: 20-50 users
   - Target demographics: families, small communities
   - Mix of Indonesia/international users
2. Feedback collection:
   ```typescript
   // In-app feedback button (Settings → Kirim Masukan)
   const FeedbackScreen: React.FC = () => {
     return (
       <ScrollView style={styles.container}>
         <Text style={styles.title}>{t('feedback.title')}</Text>

         <SegmentedControl
           options={[
             { label: 'Bug', value: 'bug' },
             { label: 'Saran', value: 'feature' },
             { label: 'Lainnya', value: 'other' },
           ]}
           selected={type}
           onSelect={setType}
         />

         <TextInput
           style={styles.input}
           placeholder={t('feedback.describePlaceholder')}
           multiline
           numberOfLines={5}
           value={description}
           onChangeText={setDescription}
         />

         <Button
           title={t('feedback.attachScreenshot')}
           variant="outline"
           icon="camera"
           onPress={captureScreenshot}
         />

         <Button
           title={t('feedback.send')}
           onPress={submitFeedback}
           loading={isSubmitting}
         />
       </ScrollView>
     );
   };
   ```
3. Crash reporting:
   - Setup Firebase Crashlytics
   - Non-fatal error logging
   - Crash-free rate target: > 99%
4. Analytics (basic):
   - DAU/MAU tracking
   - Feature usage: chat, documents, entities
   - Error rate per endpoint
   - App open count
5. Beta testing checklist:
   ```
   Day 1-3: Core flow testing
   - [ ] Register with phone number
   - [ ] Create personal chat
   - [ ] Send text messages
   - [ ] Create group
   - [ ] Send group messages
   
   Day 4-7: Feature testing
   - [ ] Create topics
   - [ ] Create documents (all templates)
   - [ ] Use block editor
   - [ ] Lock/sign documents
   - [ ] Add entities
   - [ ] Search messages/documents
   
   Day 8-10: Edge cases
   - [ ] Offline messaging
   - [ ] Large groups (50+ messages)
   - [ ] Long documents (100+ blocks)
   - [ ] Arabic language + RTL
   - [ ] Background/foreground transitions
   - [ ] Push notifications
   ```

### Acceptance Criteria:
- [ ] Tester group: 20+ users
- [ ] In-app feedback mechanism
- [ ] Crashlytics configured
- [ ] Basic analytics tracking
- [ ] Testing checklist distributed
- [ ] Bug reports triaged within 24 hours

### Testing:
- [ ] Feedback submission works
- [ ] Crashes reported to Crashlytics
- [ ] Analytics events recorded

---

## Task 26.4: Bug Fixing & Stabilization

**Input:** Task 26.3 feedback
**Output:** Stable beta release

### Steps:
1. Bug triage priority:
   ```
   P0 (Critical): App crash, data loss, auth broken
   → Fix within 24 hours

   P1 (High): Feature broken, incorrect behavior
   → Fix within 3 days

   P2 (Medium): UI issues, minor functionality
   → Fix before production release

   P3 (Low): Polish, edge cases
   → Fix if time permits
   ```
2. Fix → test → re-release cycle:
   - Fix bugs on `develop` branch
   - Run full test suite
   - Tag new beta (beta-v0.1.1, beta-v0.1.2, etc.)
   - Auto-deploy to TestFlight / Play Console
   - Notify testers
3. Stabilization criteria for production:
   - Crash-free rate: > 99%
   - All P0/P1 bugs resolved
   - No P2 bugs in core flows
   - All E2E tests passing
   - Performance targets met
   - At least 1 week without P0/P1

### Acceptance Criteria:
- [ ] All P0 bugs fixed
- [ ] All P1 bugs fixed
- [ ] P2 bugs reviewed and prioritized
- [ ] Crash-free rate > 99%
- [ ] 1 week stability period passed
- [ ] Tester approval from majority

### Testing:
- [ ] Full E2E test suite passes
- [ ] Manual testing of all fixed bugs
- [ ] Regression testing of core flows

---

## Phase 26 Review

### Testing Checklist:
- [ ] Store setup: Apple + Google
- [ ] Beta build: iOS + Android
- [ ] Distribution: TestFlight + Play Console
- [ ] Feedback collection working
- [ ] Crash reporting active
- [ ] Bug triage process established
- [ ] Stabilization criteria met

### Review Checklist:
- [ ] App metadata complete
- [ ] Screenshots capture key features
- [ ] Privacy policy published
- [ ] Crashlytics data clean
- [ ] Commit: `release: beta v0.1.0`
